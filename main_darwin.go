// +build darwin

package main

import "os"

func MakeFile(name string, size int64) (*os.File, error) {
	return os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
}
