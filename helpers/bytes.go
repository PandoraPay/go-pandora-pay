package helpers

import (
	"math/rand"
	"unsafe"
)

const HashSize = 32
const ChecksumSize = 4

type Hash [HashSize]byte
type Checksum [ChecksumSize]byte

func ConvertHash(s []byte) (a *Hash) {
	if len(a) <= len(s) {
		a = (*Hash)(unsafe.Pointer(&s[0]))
	}
	return a
}

func Byte32(s []byte) (a *[32]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}

func Byte20(s []byte) (a *[20]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}

func Byte33(s []byte) (a *[33]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}

func Byte65(s []byte) (a *[65]byte) {
	if len(a) <= len(s) {
		a = (*[len(a)]byte)(unsafe.Pointer(&s[0]))
	}
	return a
}

func RandomBytes(size int) []byte {
	a := make([]byte, size)
	rand.Read(a)
	return a
}

func EmptyBytes(size int) []byte {
	return make([]byte, size)
}
