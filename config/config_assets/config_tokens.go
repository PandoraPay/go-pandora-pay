package config_assets

import (
	"errors"
	"math"
)

var (
	ASSETS_DECIMAL_SEPARATOR_MAX      = int(10)
	ASSETS_DECIMAL_SEPARATOR_MAX_BYTE = byte(10)
)

func AssetsConvertToUnits(number float64, decimalSeparator int) (uint64, error) {

	if decimalSeparator > ASSETS_DECIMAL_SEPARATOR_MAX {
		return 0, errors.New("DecimalSeparator is higher than was supposed")
	}
	if decimalSeparator == 0 {
		return uint64(number), nil
	}
	coinDenomination := math.Pow10(decimalSeparator)
	if number < float64(math.MaxUint64)/coinDenomination {
		return uint64(number * coinDenomination), nil
	}
	return 0, errors.New("Error converting to units")
}

func AssetsConvertToBase(number uint64, decimalSeparator int) (float64, error) {
	if decimalSeparator > ASSETS_DECIMAL_SEPARATOR_MAX {
		return 0, errors.New("DecimalSeparator is higher than was supposed")
	}
	if decimalSeparator == 0 {
		return float64(number), nil
	}
	coinDenomination := math.Pow10(decimalSeparator)
	return float64(number) / coinDenomination, nil
}
