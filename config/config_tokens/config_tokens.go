package config_tokens

import (
	"errors"
	"math"
)

var (
	TOKENS_DECIMAL_SEPARATOR_MAX      = int(10)
	TOKENS_DECIMAL_SEPARATOR_MAX_BYTE = byte(10)
)

func TokensConvertToUnits(number float64, decimalSeparator int) (uint64, error) {

	if decimalSeparator > TOKENS_DECIMAL_SEPARATOR_MAX {
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

func TokensConvertToBase(number uint64, decimalSeparator int) (float64, error) {
	if decimalSeparator > TOKENS_DECIMAL_SEPARATOR_MAX {
		return 0, errors.New("DecimalSeparator is higher than was supposed")
	}
	if decimalSeparator == 0 {
		return float64(number), nil
	}
	coinDenomination := math.Pow10(decimalSeparator)
	return float64(number) / coinDenomination, nil
}
