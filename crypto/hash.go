package crypto

import (
	"math/rand"
	"unsafe"
)

const HashSize = 32
const ChecksumSize = 4

type Hash [HashSize]byte
type Checksum [ChecksumSize]byte

func RandomHash() (hash Hash) {
	a := make([]byte, HashSize)
	rand.Read(a)
	return *ConvertHash(a)
}

func ConvertHash(s []byte) (a *Hash) {
	if len(a) <= len(s) {
		a = (*Hash)(unsafe.Pointer(&s[0]))
	}
	return a
}
