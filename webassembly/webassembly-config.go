package webassembly

import (
	"pandora-pay/config"
	"pandora-pay/config/reward"
	"strconv"
	"syscall/js"
)

func convertToUnitsUint64(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (out interface{}, err error) {
		value, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return
		}

		if value, err = config.ConvertToUnitsUint64(value); err != nil {
			return
		}
		return strconv.FormatUint(value, 10), nil
	})
}

func convertToUnits(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (out interface{}, err error) {
		value, err := strconv.ParseFloat(args[0].String(), 10)
		if err != nil {
			return
		}

		value2, err := config.ConvertToUnits(value)
		if err != nil {
			return
		}

		return strconv.FormatUint(value2, 10), nil
	})
}

func convertToBase(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (out interface{}, err error) {
		value, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return
		}

		value2 := config.ConvertToBase(value)
		return strconv.FormatFloat(value2, 'f', 10, 64), nil
	})
}

func getRewardAt(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (out interface{}, err error) {
		value := reward.GetRewardAt(uint64(args[0].Int()))
		return value, nil
	})
}
