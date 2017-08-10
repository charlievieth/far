package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/charlievieth/zero"
)

const (
	kB = 1 << (10 * (iota + 1))
	mB
	gB
)

const BufSize = 32 * 1024

type Copier struct {
	r   io.Reader
	w   *os.File
	off int64
	buf []byte
}

func NewCopier(r io.Reader, w *os.File) *Copier {
	return &Copier{
		r:   r,
		w:   w,
		buf: make([]byte, BufSize),
	}
}

func (c *Copier) Copy() (written int64, err error) {
	for {
		nr, er := c.r.Read(c.buf)
		if nr > 0 {
			nw := BufSize
			var ew error
			if !zero.Zero(c.buf[:nr]) {
				nw, ew = c.w.WriteAt(c.buf[:nr], c.off)
			}
			written += int64(nw)
			c.off += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

func ExtractFile(tr *tar.Reader, h *tar.Header, dirname string) error {
	t := time.Now()

	// expect a flat archive
	name := h.Name
	if filepath.Base(name) != name {
		return fmt.Errorf("tar: archive contains subdirectory: %s", name)
	}

	// only allow regular files
	mode := h.FileInfo().Mode()
	if !mode.IsRegular() {
		return fmt.Errorf("tar: unexpected file mode (%s): %s", name, mode)
	}

	path := filepath.Join(dirname, name)
	fmt.Fprintf(os.Stderr, "Extracting: %s to %s\n", name, path)

	f, err := MakeFile(path, h.Size+(h.Size/50)) // over allocate by %2
	if err != nil {
		return fmt.Errorf("tar: opening file (%s): %s", path, err)
	}

	// this is ugly, but whatever...
	handleErr := func(err error) error {
		f.Close()
		os.Remove(path)
		return err
	}

	w := Copier{
		r:   tr,
		w:   f,
		buf: make([]byte, BufSize),
	}
	n, err := w.Copy()
	if err != nil {
		return handleErr(err)
	}

	// we over allocated - fix size
	if err := f.Truncate(n); err != nil {
		return handleErr(err)
	}
	if err := f.Close(); err != nil {
		return handleErr(err)
	}

	d := time.Since(t)
	mbs := float64(d/mB) / d.Seconds()
	fmt.Fprintf(os.Stderr, "Extracted (%s) in: %s - %.2fMB/s\n",
		name, time.Since(t), mbs)

	return nil
}

var WorkingDirectory string

func init() {
	flag.StringVar(&WorkingDirectory, "directory", "", "Working directory")
	flag.StringVar(&WorkingDirectory, "C", "", "Working directory")
}

func main() {
	flag.Parse()
	if len(flag.Args()) != 1 {
		Fatal("USAGE [OPTIONS] ARCHIVE")
	}
	tarfile := flag.Arg(0)

	// read file into memory so we don't compete for IO
	t := time.Now()
	fmt.Fprintln(os.Stderr, "reading tar archive:", tarfile)
	b, err := ioutil.ReadFile(tarfile)
	if err != nil {
		Fatal(err)
	}
	fmt.Fprintln(os.Stderr, "done reading tar archive:", time.Since(t))

	gr, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		Fatal(err)
	}
	defer gr.Close()

	dirname, err := os.Getwd()
	if err != nil {
		Fatal(err)
	}
	if WorkingDirectory != "" {
		dirname, err = filepath.Abs(WorkingDirectory)
		if err != nil {
			Fatal(err)
		}
	}
	fmt.Fprintln(os.Stderr, "working directory:", dirname)

	// normally I'd error, but let's just make the thing
	if err := os.MkdirAll(dirname, 0744); err != nil {
		Fatal(err)
	}

	t = time.Now()
	fmt.Fprintln(os.Stderr, "starting extraction...")
	tr := tar.NewReader(gr)
	for {
		h, e := tr.Next()
		if h != nil {
			if err := ExtractFile(tr, h, dirname); err != nil {
				Fatal(err)
			}
		}
		if e != nil {
			break
		}
	}
	fmt.Fprintln(os.Stderr, "extraction complete:", time.Since(t))
	fmt.Fprintln(os.Stderr, "ain't that fast as shit - we're done here")
}

func Fatal(err interface{}) {
	if err == nil {
		return
	}
	var format string
	if _, file, line, ok := runtime.Caller(1); ok && file != "" {
		format = fmt.Sprintf("Error (%s:%d)", filepath.Base(file), line)
	} else {
		format = "Error"
	}
	switch err.(type) {
	case error, string:
		fmt.Fprintf(os.Stderr, "%s: %s\n", format, err)
	default:
		fmt.Fprintf(os.Stderr, "%s: %#v\n", format, err)
	}
	os.Exit(1)
}
