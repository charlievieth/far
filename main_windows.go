package main

import (
	"os"
	"syscall"
	"unsafe"
)

const FSCTL_SET_SPARSE = 590020

type FILE_SET_SPARSE_BUFFER struct {
	SetSparse uint8
}

func MakeFile(name string, size int64) (*os.File, error) {

	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	handle := syscall.Handle(f.Fd())

	// Make file sparse

	buf := FILE_SET_SPARSE_BUFFER{SetSparse: 1}
	var out uint32
	err = syscall.DeviceIoControl(
		handle,                        // hDevice
		FSCTL_SET_SPARSE,              // dwIoControlCode
		(*byte)(unsafe.Pointer(&buf)), // lpInBuffer
		uint32(unsafe.Sizeof(buf)),    // nInBufferSize
		nil,  // lpOutBuffer
		0,    // nOutBufferSize
		&out, // lpBytesReturned
		nil,  // lpOverlapped
	)
	if err != nil {
		f.Close()
		return nil, err
	}

	if _, err := f.Seek(size, 0); err != nil {
		f.Close()
		return nil, err
	}

	if err := syscall.SetEndOfFile(handle); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}
