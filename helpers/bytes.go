package helpers

import (
	"math/rand"
	"unsafe"
)

const HashSize = 32
const ChecksumSize = 4

type Hash [HashSize]byte
type Checksum [ChecksumSize]byte

func ConvertHash(s []byte) *Hash {
	if HashSize <= len(s) {
		return (*Hash)(unsafe.Pointer(&s[0]))
	}
	panic("invalid hash")
}

func Byte32(s []byte) *[32]byte {
	if 32 <= len(s) {
		return (*[32]byte)(unsafe.Pointer(&s[0]))
	}
	panic("invalid byte32 length")
}

func Byte20(s []byte) *[20]byte {
	if 20 <= len(s) {
		return (*[20]byte)(unsafe.Pointer(&s[0]))
	}
	panic("invalid byte20 length")
}

func Byte33(s []byte) *[33]byte {
	if 33 <= len(s) {
		return (*[33]byte)(unsafe.Pointer(&s[0]))
	}
	panic("invalid byte33 length")
}

func Byte65(s []byte) *[65]byte {
	if 65 <= len(s) {
		return (*[65]byte)(unsafe.Pointer(&s[0]))
	}
	panic("invalid byte65 length")
}

func RandomBytes(size int) []byte {
	a := make([]byte, size)
	rand.Read(a)
	return a
}

func EmptyBytes(size int) []byte {
	return make([]byte, size)
}
