package zero

import (
	"bytes"
	"testing"
)

const (
	kB int64 = 1 << (10 * (iota + 1))
	mB
)

// Byte arrays of various sizes with the last byte set to 1.
var (
	one4Kb   = [kB * 4]byte{kB*4 - 1: 1}
	one32Kb  = [kB * 32]byte{kB*32 - 1: 1}
	one128Kb = [kB * 128]byte{kB*128 - 1: 1}
	one4Mb   = [mB * 4]byte{mB*4 - 1: 1}
)

// equalZero is a duplicate of portable zero and is here for reference.

var zeroBuf [16 * 1024]byte // duplicate of portable byte array

func equalZero(b []byte) bool {
	for len(b) >= len(zeroBuf) {
		if !bytes.Equal(b[0:len(zeroBuf)], zeroBuf[0:]) {
			return false
		}
		b = b[len(zeroBuf):]
	}
	return bytes.Equal(b, zeroBuf[:len(b)])
}

func benchmarkEqual(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		if equalZero(buf) {
			b.Fatal("benchmarkEqual")
		}
	}
}

func benchmarkZero(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		if Zero(buf) {
			b.Fatal("benchmarkZero")
		}
	}
}

func BenchmarkZero_One_4Kb(b *testing.B) {
	benchmarkZero(b, one4Kb[0:])
}

func BenchmarkEqual_One_4Kb(b *testing.B) {
	benchmarkEqual(b, one4Kb[0:])
}

func BenchmarkZero_One_32Kb(b *testing.B) {
	benchmarkZero(b, one32Kb[0:])
}

func BenchmarkEqual_One_32Kb(b *testing.B) {
	benchmarkEqual(b, one32Kb[0:])
}

func BenchmarkZero_One_128Kb(b *testing.B) {
	benchmarkZero(b, one128Kb[0:])
}

func BenchmarkEqual_One_128Kb(b *testing.B) {
	benchmarkEqual(b, one128Kb[0:])
}

func BenchmarkZero_One_4Mb(b *testing.B) {
	benchmarkZero(b, one4Mb[0:])
}

func BenchmarkEqual_One_4Mb(b *testing.B) {
	benchmarkEqual(b, one4Mb[0:])
}

// Naive solution - here for reference

func slowZero(b []byte) bool {
	for _, c := range b {
		if c != 0 {
			return false
		}
	}
	return true
}

func benchmarkSlow(b *testing.B, buf []byte) {
	for i := 0; i < b.N; i++ {
		if slowZero(buf) {
			b.Fatal("benchmarkSlow")
		}
	}
}

func BenchmarkSlow_One_4Kb(b *testing.B) {
	benchmarkSlow(b, one4Kb[0:])
}

func BenchmarkSlow_One_128Kb(b *testing.B) {
	benchmarkSlow(b, one128Kb[0:])
}
