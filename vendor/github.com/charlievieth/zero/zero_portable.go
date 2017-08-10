// +build !amd64

package zero

import "bytes"

var zeroBuf [16 * 1024]byte

func Zero(b []byte) bool {
	for len(b) >= len(zeroBuf) {
		if !bytes.Equal(b[0:len(zeroBuf)], zeroBuf[0:]) {
			return false
		}
		b = b[len(zeroBuf):]
	}
	return bytes.Equal(b, zeroBuf[:len(b)])
}
