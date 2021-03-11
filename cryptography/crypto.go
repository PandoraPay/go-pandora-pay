package cryptography

import (
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"
	"math/big"
	"unsafe"
)

func SHA3Hash(b []byte) (result Hash) {
	h := sha3.New256()
	h.Write(b)
	return ConvertHash(h.Sum(nil))
}

func SHA3(b []byte) []byte {
	h := sha3.New256()
	h.Write(b)
	return h.Sum(nil)
}

func RIPEMD(b []byte) *[20]byte {
	h := ripemd160.New()
	h.Write(b)
	s := h.Sum(nil)
	if 20 <= len(s) {
		return (*[20]byte)(unsafe.Pointer(&s[0]))
	}
	panic("invalid byte20 length")
}

func GetChecksum(b []byte) Checksum {
	s := *RIPEMD(b)
	return *(*[ChecksumSize]byte)(unsafe.Pointer(&s[0]))
}

func ComputePublicKeyHash(publicKey *[33]byte) *[20]byte {
	return RIPEMD(SHA3(publicKey[:]))
}

func ComputeKernelHash(hash Hash, stakingAmount uint64) (out Hash) {

	number := new(big.Int).Div(new(big.Int).SetBytes(hash[:]), new(big.Int).SetUint64(stakingAmount))

	buf := number.Bytes()
	copy(out[HashSize-len(buf):], buf)

	return out
}
