package config

import (
	"math"
)

var BURN_PUBLIC_KEY_HASH = [20]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

var (
	DECIMAL_SEPARATOR = byte(7)
	COIN_DENOMINATION = uint64(math.Pow(10, float64(DECIMAL_SEPARATOR)))

	MAX_SUPPLY_COINS = uint64(42000000000)

	NATIVE_TOKEN_NAME        = "PANDORA"
	NATIVE_TOKEN_TICKER      = "PANDORA"
	NATIVE_TOKEN_DESCRIPTION = "PANDORA NATIVE TOKEN"

	NATIVE_TOKEN      = []byte{}
	NATIVE_TOKEN_FULL = [20]byte{}
)

func ConvertToUnits(number uint64) uint64 {
	return number * COIN_DENOMINATION
}

func ConvertToBase(number uint64) uint64 {
	return number / COIN_DENOMINATION
}
