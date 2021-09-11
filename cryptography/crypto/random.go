package crypto

import (
	"crypto/rand"
	"math/big"
	"pandora-pay/cryptography/bn256"
)

func RandomScalar() *big.Int {

	for {
		a, _ := rand.Int(rand.Reader, bn256.Order)
		if a.Sign() > 0 {
			return a
		}

	}
}

// this will return fixed random scalar
func RandomScalarFixed() *big.Int {
	//return new(big.Int).Set(fixed)

	return RandomScalar()
}

type KeyPair struct {
	x *big.Int  // secret key
	y *bn256.G1 // public key
}
