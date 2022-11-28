package config_stake

import (
	"pandora-pay/config/arguments"
	"pandora-pay/config/config_coins"
)

func GetRequiredStake(blockHeight uint64) (requiredStake uint64) {

	var err error

	if requiredStake, err = config_coins.ConvertToUnitsUint64(100); err != nil {
		panic(err)
	}

	return
}

func GetPendingStakeWindow(blockHeight uint64) uint64 {

	if arguments.Arguments["--new-devnet"] == true {

		if blockHeight == 0 {
			return 1
		}
		return 10
	}

	return 60
}
