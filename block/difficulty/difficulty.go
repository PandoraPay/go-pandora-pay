package difficulty

import (
	"math/big"
	"pandora-pay/crypto"
)

var (
	bigZero = big.NewInt(0)

	bigOne = big.NewInt(1)

	oneMAX256 = new(big.Int).Lsh(bigOne, 256) // 0xFFFFFFFF....

)

func HashToBig(buf crypto.Hash) *big.Int {

	// little-endian to big-endian
	blen := crypto.HashSize
	for i := 0; i < blen/2; i++ {
		buf[i], buf[blen-1-i] = buf[blen-1-i], buf[i]
	}

	return new(big.Int).SetBytes(buf[:])
}

// this function calculates the difficulty in big num form
func ConvertDifficultyToBig(difficulty uint64) *big.Int {
	if difficulty == 0 {
		panic("difficulty can never be zero")
	}
	// (1 << 256) / (difficultyNum )
	difficultyInt := new(big.Int).SetUint64(difficulty)
	denominator := new(big.Int).Add(difficultyInt, bigZero) // above 2 lines can be merged
	return new(big.Int).Div(oneMAX256, denominator)
}

func CheckKernelHashBig(kernelHash crypto.Hash, difficulty *big.Int) bool {

	bigKernelHash := HashToBig(kernelHash)

	if bigKernelHash.Cmp(difficulty) <= 0 {
		return true
	}
	return false
}
