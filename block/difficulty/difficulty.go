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

func ConvertIntegerDifficultyToBig(difficulty *big.Int) *big.Int {

	if difficulty.Cmp(bigZero) == 0 {
		panic("Difficulty can never be zero. Division by zero")
	}

	return new(big.Int).Div(oneMAX256, difficulty)
}

func CheckKernelHashBig(kernelHash crypto.Hash, difficulty *big.Int) bool {

	bigKernelHash := HashToBig(kernelHash)

	bigDifficulty := ConvertIntegerDifficultyToBig(difficulty)
	if bigKernelHash.Cmp(bigDifficulty) <= 0 {
		return true
	}
	return false
}
