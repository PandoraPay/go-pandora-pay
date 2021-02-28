package stake

import "pandora-pay/config"

func GetRequiredStake(blockHeight uint64) (stake uint64) {

	if blockHeight == 0 {
		stake = 0
	} else {
		stake = 100
	}

	return config.ConvertToUnits(stake)
}
