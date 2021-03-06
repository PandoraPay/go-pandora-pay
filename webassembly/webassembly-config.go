package webassembly

import (
	"pandora-pay/config"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_tokens"
	"strconv"
	"syscall/js"
)

func convertToUnitsUint64(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {
		value, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		if value, err = config.ConvertToUnitsUint64(value); err != nil {
			return nil, err
		}
		return strconv.FormatUint(value, 10), nil
	})
}

func convertToUnits(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {
		number, err := strconv.ParseFloat(args[0].String(), 10)
		if err != nil {
			return nil, err
		}

		value2, err := config.ConvertToUnits(number)
		if err != nil {
			return nil, err
		}

		return strconv.FormatUint(value2, 10), nil
	})
}

func convertToBase(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {
		number, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		value2 := config.ConvertToBase(number)
		return strconv.FormatFloat(value2, 'f', config.DECIMAL_SEPARATOR, 64), nil
	})
}

func tokensConvertToUnits(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {
		number, err := strconv.ParseFloat(args[0].String(), 10)
		if err != nil {
			return nil, err
		}

		decimalSeparator, err := strconv.Atoi(args[1].String())
		if err != nil {
			return nil, err
		}

		value2, err := config_tokens.TokensConvertToUnits(number, decimalSeparator)
		if err != nil {
			return nil, err
		}

		return strconv.FormatUint(value2, 10), nil
	})
}

func tokensConvertToBase(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {

		number, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		decimalSeparator, err := strconv.Atoi(args[1].String())
		if err != nil {
			return nil, err
		}

		value2, err := config_tokens.TokensConvertToBase(number, decimalSeparator)
		if err != nil {
			return nil, err
		}

		return strconv.FormatFloat(value2, 'f', decimalSeparator, 64), nil
	})
}

func getRewardAt(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {
		value := config_reward.GetRewardAt(uint64(args[0].Int()))
		return value, nil
	})
}
