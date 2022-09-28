package config

import "pandora-pay/config/config_coins"

var (
	HCAPTCHA_SECRET_KEY        = ""
	FAUCET_TESTNET_ENABLED     = false
	FAUCET_TESTNET_COINS       = uint64(100)
	FAUCET_TESTNET_COINS_UNITS = config_coins.ConvertToUnitsUint64Forced(FAUCET_TESTNET_COINS)
)
