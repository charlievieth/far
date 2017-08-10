package main

import (
	"os"
	"syscall"
)

func MakeFile(name string, size int64) (*os.File, error) {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	if _, err := f.Seek(size, 0); err != nil {
		f.Close()
		return nil, err
	}

	if err := syscall.SetEndOfFile(syscall.Handle(f.Fd())); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}
