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
