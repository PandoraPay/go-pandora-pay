package cryptography

import (
	"math/rand"
)

const HashSize = 32
const KeyHashSize = 20
const ChecksumSize = 4

func RandomHash() (hash []byte) {
	a := make([]byte, HashSize)
	rand.Read(a)
	return a
}
