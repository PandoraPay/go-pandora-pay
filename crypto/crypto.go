package crypto

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"pandora-pay/helpers"
)

func SHA3Hash(b []byte) (result helpers.Hash) {
	h := sha3.New256()
	h.Write(b)
	return *helpers.ConvertHash(h.Sum(nil))
}

func SHA3(b []byte) []byte {
	h := sha3.New256()
	h.Write(b)
	return h.Sum(nil)
}

func RIPEMD(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}

func ComputePublicKeyHash(publicKey []byte) []byte {
	return RIPEMD(SHA3(publicKey))
}
