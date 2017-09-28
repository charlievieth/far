package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/charlievieth/zero"
)

const (
	kB = 1 << (10 * (iota + 1))
	mB
	gB

	BufSize = 32 * mB
)

var Debugf = func(format string, a ...interface{}) {}

func init() {
	const usageMsg = "Usage %s: ARCHIVE [-C DIR]\n"
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageMsg, filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func parseFlags() (archive, dirname string, _ error) {
	var workingDirectory string
	var enableDebug bool

	flag.StringVar(&workingDirectory, "directory", "", "change to directory DIR")
	flag.StringVar(&workingDirectory, "C", "", "change to directory DIR")
	flag.BoolVar(&enableDebug, "debug", false, "print debugging information")

	flag.Parse()

	if enableDebug {
		Debugf = log.New(os.Stderr, "debug: ", 0).Printf
		Debugf("DEBUG output enabled")
	}

	Debugf("validating command line arguments")

	switch flag.NArg() {
	case 0:
		return "", "", errors.New("missing ARCHIVE argument")
	case 1:
		archive = flag.Arg(0)
		fi, err := os.Stat(archive)
		if err != nil {
			return "", "", fmt.Errorf("invalid ARCHIVE argument (%s): %s", archive, err)
		}
		if fi.IsDir() {
			return "", "", fmt.Errorf("invalid ARCHIVE argument (%s): invalid file type",
				archive)
		}
	default:
		return "", "", fmt.Errorf("too many arguments: %s", flag.Args())
	}

	if workingDirectory != "" {
		Debugf("validating [C|directory] argument: %s", workingDirectory)
		fi, err := os.Stat(workingDirectory)
		if err != nil {
			return "", "", fmt.Errorf("invalid argument [C|directory] (%s): %s",
				workingDirectory, err)
		}
		if !fi.IsDir() {
			return "", "", fmt.Errorf("invalid argument [C|directory] (%s): not a diretory",
				workingDirectory)
		}
		dirname = workingDirectory
	} else {
		pwd, err := os.Getwd()
		if err != nil {
			return "", "", err
		}
		Debugf("no [C|directory] argment provided using: %s", pwd)
		dirname = pwd
	}

	Debugf("Argument [ARCHIVE]: %s", archive)
	Debugf("Argument [C|dirname]: %s", dirname)

	return archive, dirname, nil
}

func Copy(dst *os.File, src io.Reader, buf []byte) (written int64, err error) {
	var off int64
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw := BufSize
			var ew error
			if !zero.Zero(buf[:nr]) {
				nw, ew = dst.WriteAt(buf[:nr], off)
			}
			written += int64(nw)
			off += int64(nw)
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
	Debugf("Extracting: %s to %s\n", name, path)

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

	n, err := Copy(f, tr, make([]byte, BufSize))
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
	Debugf("Extracted (%s) in: %s - %.2fMB/s\n", name, time.Since(t), mbs)

	return nil
}

func realMain() error {
	tarfile, dirname, err := parseFlags()
	if err != nil {
		flag.Usage()
		return errors.New("invalid command line arguments")
	}

	fa, err := os.Open(tarfile)
	if err != nil {
		return err
	}
	defer fa.Close()

	// Read file in chunks to improve IO
	gr, err := gzip.NewReader(bufio.NewReaderSize(fa, BufSize))
	if err != nil {
		return err
	}
	defer gr.Close()

	// normally I'd error, but let's just make the thing
	if err := os.MkdirAll(dirname, 0744); err != nil {
		return err
	}

	t := time.Now()
	Debugf("starting extraction...")
	tr := tar.NewReader(gr)
	for {
		h, e := tr.Next()
		if h != nil {
			if err := ExtractFile(tr, h, dirname); err != nil {
				return err
			}
		}
		if e != nil {
			break
		}
	}

	Debugf("extraction complete:", time.Since(t))
	return nil
}

func main() {
	if err := realMain(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		fmt.Fprintln(os.Stderr, "Error is not recoverable: exiting now")
		os.Exit(1)
	}
}
