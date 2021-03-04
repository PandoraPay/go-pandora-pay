package cryptography

import (
	"math/rand"
	"pandora-pay/helpers"
)

func RandomHash() (hash helpers.Hash) {
	a := make([]byte, helpers.HashSize)
	rand.Read(a)
	return *helpers.ConvertHash(a)
}
