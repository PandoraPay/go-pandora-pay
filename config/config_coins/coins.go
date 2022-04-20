package config_coins

import (
	"encoding/base64"
	"errors"
	"math"
	"pandora-pay/cryptography"
)

var BURN_PUBLIC_KEY = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xd, 0xe, 0xa, 0xd}

const (
	ASSET_LENGTH = cryptography.PublicKeyHashSize
)

var (
	DECIMAL_SEPARATOR       = 6
	COIN_DENOMINATION       = uint64(math.Pow10(DECIMAL_SEPARATOR))
	COIN_DENOMINATION_FLOAT = float64(math.Pow10(DECIMAL_SEPARATOR))

	MAX_SUPPLY_COINS       = uint64(42000000000)
	MAX_SUPPLY_COINS_UNITS = ConvertToUnitsUint64Forced(MAX_SUPPLY_COINS)

	NATIVE_ASSET_NAME           = "WEBD"
	NATIVE_ASSET_TICKER         = "WEBD"
	NATIVE_ASSET_IDENTIFICATION = "WEBD"
	NATIVE_ASSET_DESCRIPTION    = "WEBDOLLAR"

	NATIVE_ASSET_FULL               = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	NATIVE_ASSET_FULL_STRING        = string(NATIVE_ASSET_FULL)
	NATIVE_ASSET_FULL_STRING_BASE64 = base64.StdEncoding.EncodeToString(NATIVE_ASSET_FULL)
)

func ConvertToUnitsUint64Forced(number uint64) uint64 {
	out, err := ConvertToUnitsUint64(number)
	if err != nil {
		panic(err)
	}
	return out
}

func ConvertToUnitsUint64(number uint64) (uint64, error) {
	if number < math.MaxUint64/COIN_DENOMINATION {
		return number * COIN_DENOMINATION, nil
	}
	return 0, errors.New("Error converting to units")
}

func ConvertToUnits(number float64) (uint64, error) {
	if number < float64(math.MaxUint64)/COIN_DENOMINATION_FLOAT {
		return uint64(number * COIN_DENOMINATION_FLOAT), nil
	}
	return 0, errors.New("Error converting to units")
}

func ConvertToBase(number uint64) float64 {
	return float64(number) / COIN_DENOMINATION_FLOAT
}
