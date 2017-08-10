// +build !windows !darwin

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
	if err := syscall.Fallocate(int(f.Fd()), 0, 0, size); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}
