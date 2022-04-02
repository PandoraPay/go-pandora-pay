package cryptography

import (
	"crypto/sha256"
	"errors"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"math/big"
)

func SHA3(b []byte) []byte {
	h := sha3.New256()
	h.Write(b)
	return h.Sum(nil)
}

func SHA256(b []byte) []byte {
	h := sha256.New()
	h.Write(b)
	return h.Sum(nil)
}

func RIPEMD(b []byte) []byte {
	h := ripemd160.New()
	h.Write(b)
	return h.Sum(nil)
}

func GetChecksum(b []byte) []byte {
	return SHA256(SHA256(b))[:ChecksumSize]
}

func GetPublicKeyHash(publicKey []byte) []byte {
	return RIPEMD(SHA256(publicKey))
}

func ComputeKernelHash(hash []byte, stakingAmount uint64) ([]byte, error) {

	if stakingAmount == 0 {
		return nil, errors.New("Staking Amount can not be zero")
	}

	number := new(big.Int).Div(new(big.Int).SetBytes(hash), new(big.Int).SetUint64(stakingAmount))

	buf := number.Bytes()
	var out [HashSize]byte
	copy(out[HashSize-len(buf):], buf[:])

	return out[:], nil
}
