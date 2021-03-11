package cryptography

import (
	"encoding/hex"
	"math/rand"
	"unsafe"
)

const HashSize = 32
const ChecksumSize = 4

type Hash [HashSize]byte
type Checksum [ChecksumSize]byte

func (h *Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (c *Checksum) String() string {
	return hex.EncodeToString(c[:])
}

func ConvertHash(s []byte) *Hash {
	if HashSize <= len(s) {
		return (*Hash)(unsafe.Pointer(&s[0]))
	}
	panic("invalid hash")
}

func RandomHash() (hash Hash) {
	a := make([]byte, HashSize)
	rand.Read(a)
	return *ConvertHash(a)
}
