package difficulty

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

var (
	DIFFICULTY_MIN_CHANGE_FACTOR = new(big.Float).SetFloat64(0.5)
	DIFFICULTY_MAX_CHANGE_FACTOR = new(big.Float).SetFloat64(2)
)

func ConvertHashToDifficulty(hash []byte) *big.Int {
	return new(big.Int).Div(config.BIG_INT_MAX_256, new(big.Int).SetBytes(hash))
}

func ConvertTargetToDifficulty(target *big.Int) *big.Int {
	return new(big.Int).Div(config.BIG_INT_MAX_256, target)
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

func CheckKernelHashBig(kernelHash []byte, difficulty *big.Int) bool {
	return new(big.Int).SetBytes(kernelHash).Cmp(difficulty) <= 0
}

func NextTargetBig(deltaTotalDifficulty *big.Int, deltaTime uint64) (*big.Int, error) {

	expectedTime := config.BLOCK_TIME * config.DIFFICULTY_BLOCK_WINDOW

	change := new(big.Float).Quo(new(big.Float).SetUint64(deltaTime), new(big.Float).SetUint64(expectedTime))

	if change.Cmp(DIFFICULTY_MIN_CHANGE_FACTOR) < 0 {
		change = DIFFICULTY_MIN_CHANGE_FACTOR
	}
	if change.Cmp(DIFFICULTY_MAX_CHANGE_FACTOR) > 0 {
		change = DIFFICULTY_MAX_CHANGE_FACTOR
	}

	gui.Log("deltaTotalDifficulty", deltaTotalDifficulty.String())
	gui.Log(strconv.FormatUint(deltaTime, 10) + "  expected " + strconv.FormatUint(expectedTime, 10))
	gui.Log("change " + change.String())

	averageDifficulty := new(big.Float).Quo(new(big.Float).SetInt(deltaTotalDifficulty), new(big.Float).SetUint64(config.DIFFICULTY_BLOCK_WINDOW))
	averageTarget := new(big.Float).Quo(config.BIG_FLOAT_MAX_256, averageDifficulty)

	newTarget := new(big.Float).Mul(averageTarget, change)

	gui.Log("before " + averageTarget.String())
	gui.Log("after " + newTarget.String())
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

	hexstr := hex.EncodeToString(final.Bytes())
	gui.Log("final "+hex.EncodeToString(helpers.EmptyBytes(32-len(hexstr)/2))+hexstr, final.String())

	return final, nil
}
