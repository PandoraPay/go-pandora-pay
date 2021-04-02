package config

import (
	"errors"
	"math"
)

var BURN_PUBLIC_KEY_HASH = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xd, 0xe, 0xa, 0xd}

var (
	DECIMAL_SEPARATOR       = int(7)
	COIN_DENOMINATION       = uint64(math.Pow10(DECIMAL_SEPARATOR))
	COIN_DENOMINATION_FLOAT = float64(math.Pow10(DECIMAL_SEPARATOR))

	MAX_SUPPLY_COINS = uint64(42000000000)

	TOKEN_LENGTH = 20

	NATIVE_TOKEN_NAME        = "PANDORA"
	NATIVE_TOKEN_TICKER      = "PANDORA"
	NATIVE_TOKEN_DESCRIPTION = "PANDORA NATIVE TOKEN"

	NATIVE_TOKEN        = []byte{}
	NATIVE_TOKEN_STRING = string(NATIVE_TOKEN)
	NATIVE_TOKEN_FULL   = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

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
