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

type CryptoRandSource struct{}

func NewCryptoRandSource() CryptoRandSource {
	return CryptoRandSource{}
}

func (_ CryptoRandSource) Int63() int64 {
	var b [8]byte
	rand.Read(b[:])
	// mask off sign bit to ensure positive number
	return int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1))
}

func (_ CryptoRandSource) Seed(_ int64) {}

func ShuffleArray(count int) []int {

	array := make([]int, count)
	for i := 0; i < count; i++ {
		array[i] = i
	}

	Global_Random.Shuffle(count, func(i, j int) {
		array[i], array[j] = array[j], array[i]
	})

	return array
}

func ShuffleArray_for_Zether(count int) []int {

	for {
		witness_index := ShuffleArray(count)

		// make sure sender and receiver are not both odd or both even
		// sender will always be at  witness_index[0] and receiver will always be at witness_index[1]
		if witness_index[0]%2 != witness_index[1]%2 {
			return witness_index
		}
	}

}

var Global_Random = rand.New(NewCryptoRandSource())
