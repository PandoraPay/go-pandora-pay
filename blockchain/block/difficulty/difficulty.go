package difficulty

import (
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

var (
	DIFFICULTY_MIN_CHANGE_FACTOR = new(big.Float).SetFloat64(0.5)
	DIFFICULTY_MAX_CHANGE_FACTOR = new(big.Float).SetFloat64(2)
)

func HashToBig(buf helpers.Hash) *big.Int {

	// little-endian to big-endian
	blen := helpers.HashSize
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

func CheckKernelHashBig(kernelHash helpers.Hash, difficulty *big.Int) bool {

	bigKernelHash := HashToBig(kernelHash)

	if bigKernelHash.Cmp(difficulty) <= 0 {
		return true
	}
	return false
}

func NextDifficultyBig(deltaTotalDifficulty *big.Int, deltaTime uint64) (*big.Int, error) {

	expectedTime := config.BLOCK_TIME * config.DIFFICULTY_BLOCK_WINDOW

	change := new(big.Float).Quo(new(big.Float).SetUint64(deltaTime), new(big.Float).SetUint64(expectedTime))

	if change.Cmp(DIFFICULTY_MIN_CHANGE_FACTOR) < 0 {
		change = DIFFICULTY_MIN_CHANGE_FACTOR
	}
	if change.Cmp(DIFFICULTY_MAX_CHANGE_FACTOR) > 0 {
		change = DIFFICULTY_MAX_CHANGE_FACTOR
	}

	averageDifficulty := new(big.Float).Quo(new(big.Float).SetInt(deltaTotalDifficulty), new(big.Float).SetUint64(config.DIFFICULTY_BLOCK_WINDOW))
	averageTarget := new(big.Float).Quo(config.BIG_FLOAT_MAX_256, averageDifficulty)

	newTarget := new(big.Float).Mul(averageTarget, change)

	str := fmt.Sprintf("%.0f", newTarget)
	final, success := new(big.Int).SetString(str, 10)
	if success == false {
		return nil, errors.New("Error rounding new target")
	}

	if final.Cmp(config.BIG_INT_ZERO) < 0 {
		final = config.BIG_INT_ONE
	}

	if final.Cmp(config.BIG_INT_MAX_256) > 0 {
		final = config.BIG_INT_MAX_256
	}

	return final, nil
}
