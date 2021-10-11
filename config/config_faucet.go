package config

var (
	HCAPTCHA_SITE_KEY          = ""
	HCAPTCHA_SECRET_KEY        = ""
	FAUCET_TESTNET_ENABLED     = false
	FAUCET_TESTNET_COINS       = uint64(100)
	FAUCET_TESTNET_COINS_UNITS = ConvertToUnitsUint64Forced(FAUCET_TESTNET_COINS)
)
