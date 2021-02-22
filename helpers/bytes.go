package helpers

import (
	"math/rand"
	"unsafe"
)

func Byte32(s []byte) (a *[32]byte) {
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
