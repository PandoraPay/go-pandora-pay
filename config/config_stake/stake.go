package config_stake

import (
	"pandora-pay/config/config_coins"
	"pandora-pay/config/globals"
)

const DELEGATING_STAKING_FEE_MAX_VALUE = uint64(10000)

func GetRequiredStake(blockHeight uint64) (requiredStake uint64) {

	var err error

	if requiredStake, err = config_coins.ConvertToUnitsUint64(100); err != nil {
		panic(err)
	}

	return
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

func GetPendingUnstakeWindow(blockHeight uint64) uint64 {

	if globals.Arguments["--new-devnet"] == true {

		if blockHeight == 0 {
			return 1
		}
		return 10
	}

	return 60
}
