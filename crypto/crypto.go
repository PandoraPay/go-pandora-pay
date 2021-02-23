package crypto

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"pandora-pay/crypto/ecdsa"
)

func SHA3Hash(b []byte) (result Hash) {
	h := sha3.New256()
	h.Write(b)
	copy(result[:], h.Sum(nil))
	return
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

func ComputePublicKey(key []byte) ([]byte, error) {

	privateKey, err := ecdsa.ToECDSA(key)
	if err != nil {
		return nil, err
	}

	return ecdsa.CompressPubkey(&privateKey.PublicKey), nil
}
