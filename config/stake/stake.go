package stake

import "pandora-pay/config"

func GetRequiredStake(blockHeight uint64) uint64 {
	if blockHeight == 0 {
		return config.ConvertToUnits(0)
	} else {
		return config.ConvertToUnits(100)
	}
}

func GetPendingStakeWindow(blockHeight uint64) uint64 {
	if blockHeight == 0 {
		return 1
	}
	return 20
}

func GetUnstakeWindow(blockHeight uint64) uint64 {
	if blockHeight < 10000 {
		return 5
	} else {
		return 5000
	}
}
