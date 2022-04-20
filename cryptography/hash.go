package cryptography

import (
	"crypto/ed25519"
	"math/rand"
)

type Bytes []byte

const HashSize = 32

const PrivateKeySize = ed25519.PrivateKeySize
const SeedSize = 64
const PublicKeySize = ed25519.PublicKeySize
const SignatureSize = ed25519.SignatureSize

const RipemdSize = 20
const PublicKeyHashSize = RipemdSize
const ChecksumSize = 4

func RandomHash() (hash []byte) {
	a := make([]byte, HashSize)
	rand.Read(a)
	return a
}
