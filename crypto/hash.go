package crypto

import "math/rand"

const HashSize = 32
const ChecksumSize = 4

type Hash [HashSize]byte
type Checksum [ChecksumSize]byte

func RandomHash() (hash Hash) {
	a := make([]byte, HashSize)
	rand.Read(a)

	copy(hash[:], a)
	return
}
