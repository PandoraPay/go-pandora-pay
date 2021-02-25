package difficulty

import (
	"math/big"
	"pandora-pay/config"
	"pandora-pay/crypto"
)

var (
	DIFFICULTY_MIN_CHANGE_FACTOR = new(big.Float).SetFloat64(0.5)
	DIFFICULTY_MAX_CHANGE_FACTOR = new(big.Float).SetFloat64(2)
)

func HashToBig(buf crypto.Hash) *big.Int {

	// little-endian to big-endian
	blen := crypto.HashSize
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

func ConvertDifficultyBigToUInt64(difficulty *big.Int) uint64 {

	if difficulty.Cmp(config.BIG_INT_ZERO) == 0 {
		panic("difficulty can never be zero")
	}

	return new(big.Int).Div(config.BIG_INT_MAX_256, difficulty).Uint64()
}

// this function calculates the difficulty in big num form
func ConvertDifficultyToBig(difficulty uint64) *big.Int {
	if difficulty == 0 {
		panic("difficulty can never be zero")
	}
	// (1 << 256) / (difficultyNum )
	difficultyInt := new(big.Int).SetUint64(difficulty)
	return new(big.Int).Div(config.BIG_INT_MAX_256, difficultyInt)
}

func CheckKernelHashBig(kernelHash crypto.Hash, difficulty *big.Int) bool {

	bigKernelHash := HashToBig(kernelHash)

	if bigKernelHash.Cmp(difficulty) <= 0 {
		return true
	}
	return false
}
