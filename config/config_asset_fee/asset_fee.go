package config_asset_fee

import "pandora-pay/config/config_coins"

func GetRequiredAssetFee(blockHeight uint64) (requiredAssetFee uint64) {

	var err error

	if requiredAssetFee, err = config_coins.ConvertToUnitsUint64(100); err != nil {
		panic(err)
	}

	return
}
