package crypto

import "math/rand"

type Hash [32]byte
type Checksum [4]byte

func RandomHash() (hash Hash) {
	a := make([]byte, 32)
	rand.Read(a)

	copy(hash[:], a)
	return
}
