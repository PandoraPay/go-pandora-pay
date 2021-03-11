package stake

import (
	"pandora-pay/config"
	"pandora-pay/config/globals"
)

func GetRequiredStake(blockHeight uint64) uint64 {
	if blockHeight == 0 {
		return config.ConvertToUnits(0)
	} else {
		return config.ConvertToUnits(100)
	}
}

func GetPendingStakeWindow(blockHeight uint64) uint64 {

	if globals.Arguments["--new-devnet"] == true {

		if blockHeight == 0 {
			return 1
		}

		return 10
	}

	return 60
}

func GetUnstakeWindow(blockHeight uint64) uint64 {

	if globals.Arguments["--new-devnet"] == true {
		return 10
	}

	return 5000
}
