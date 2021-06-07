package helpers

import (
	"encoding/hex"
	"math/rand"
)

func RandomBytes(size int) []byte {
	a := make([]byte, size)
	rand.Read(a)
	return a
}

func EmptyBytes(size int) []byte {
	return make([]byte, size)
}

func CloneBytes(a []byte) []byte {
	if a == nil {
		return nil
	}
	out := make([]byte, len(a))
	copy(out, a)
	return out
}

func DecodeHex(a string) []byte {
	out, err := hex.DecodeString(a)
	if err != nil {
		panic(err)
	}
	return out
}
