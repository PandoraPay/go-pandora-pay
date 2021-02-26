package config

import "math"

var (
	DECIMAL_SEPARATOR = 7
	COIN_DENOMINATION = uint64(math.Pow(10, float64(DECIMAL_SEPARATOR)))

	MAX_SUPPLY_COINS = 42000000000

	NATIVE_CURRENCY_NAME = "PANDORA"
	NATIVE_CURRENCY      = []byte{}
)

func ConvertToUnits(number uint64) uint64 {
	return number * COIN_DENOMINATION
}

func ConvertToBase(number uint64) uint64 {
	return number / COIN_DENOMINATION
}
