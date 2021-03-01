package helpers

import (
	"encoding/binary"
	"math/rand"
)

func RandomUint64() uint64 {
	buf := make([]byte, 8)
	rand.Read(buf) // Always succeeds, no need to check error
	return binary.LittleEndian.Uint64(buf)
}
