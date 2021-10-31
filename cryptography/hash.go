package cryptography

import (
	"math/rand"
)

type Bytes []byte

const HashSize = 32
const PrivateKeySize = 33
const PublicKeySize = 33
const RipemdSize = 20
const SignatureSize = 64
const ChecksumSize = 4

func RandomHash() (hash []byte) {
	a := make([]byte, HashSize)
	rand.Read(a)
	return a
}
