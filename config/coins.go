package config

import (
	"encoding/hex"
	"errors"
	"math"
)

var BURN_PUBLIC_KEY = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xd, 0xe, 0xa, 0xd}

var (
	DECIMAL_SEPARATOR       = 7
	COIN_DENOMINATION       = uint64(math.Pow10(DECIMAL_SEPARATOR))
	COIN_DENOMINATION_FLOAT = float64(math.Pow10(DECIMAL_SEPARATOR))

	MAX_SUPPLY_COINS       = uint64(42000000000)
	MAX_SUPPLY_COINS_UNITS = ConvertToUnitsUint64Forced(MAX_SUPPLY_COINS)

	TOKEN_LENGTH = 20

	NATIVE_TOKEN_NAME        = "PANDORA"
	NATIVE_TOKEN_TICKER      = "PANDORA"
	NATIVE_TOKEN_DESCRIPTION = "PANDORA NATIVE TOKEN"

	NATIVE_TOKEN                 = []byte{}
	NATIVE_TOKEN_STRING          = string(NATIVE_TOKEN)
	NATIVE_TOKEN_FULL            = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	NATIVE_TOKEN_FULL_STRING     = string(NATIVE_TOKEN_FULL)
	NATIVE_TOKEN_FULL_STRING_HEX = hex.EncodeToString(NATIVE_TOKEN_FULL)
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
