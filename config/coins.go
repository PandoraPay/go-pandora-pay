package config

import (
	"math"
)

var BURN_PUBLIC_KEY_HASH = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xd, 0xe, 0xa, 0xd}

var (
	DECIMAL_SEPARATOR = byte(7)
	COIN_DENOMINATION = uint64(math.Pow(10, float64(DECIMAL_SEPARATOR)))

	MAX_SUPPLY_COINS = uint64(42000000000)

	TOKEN_LENGTH = 20

	NATIVE_TOKEN_NAME        = "PANDORA"
	NATIVE_TOKEN_TICKER      = "PANDORA"
	NATIVE_TOKEN_DESCRIPTION = "PANDORA NATIVE TOKEN"

	NATIVE_TOKEN        = []byte{}
	NATIVE_TOKEN_STRING = string(NATIVE_TOKEN)
	NATIVE_TOKEN_FULL   = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
)

func ConvertToUnits(number uint64) uint64 {
	return number * COIN_DENOMINATION
}

func ConvertToBase(number uint64) uint64 {
	return number / COIN_DENOMINATION
}
