package main

import (
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/config"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_reward"
	"pandora-pay/config/config_stake"
	"strconv"
	"syscall/js"
)

func convertToUnitsUint64(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		value, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		if value, err = config_coins.ConvertToUnitsUint64(value); err != nil {
			return nil, err
		}
		return strconv.FormatUint(value, 10), nil
	})
}

func convertToUnits(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		number, err := strconv.ParseFloat(args[0].String(), 10)
		if err != nil {
			return nil, err
		}

		value2, err := config_coins.ConvertToUnits(number)
		if err != nil {
			return nil, err
		}

		return strconv.FormatUint(value2, 10), nil
	})
}

func convertToBase(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		number, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		value2 := config_coins.ConvertToBase(number)
		return strconv.FormatFloat(value2, 'f', config_coins.DECIMAL_SEPARATOR, 64), nil
	})
}

func assetsConvertToUnits(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		number, err := strconv.ParseFloat(args[0].String(), 10)
		if err != nil {
			return nil, err
		}

		decimalSeparator := args[1].Int()

		value2, err := config_assets.AssetsConvertToUnits(number, decimalSeparator)
		if err != nil {
			return nil, err
		}

		return strconv.FormatUint(value2, 10), nil
	})
}

func assetsConvertToBase(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		number, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		decimalSeparator := args[1].Int()

		value2, err := config_assets.AssetsConvertToBase(number, decimalSeparator)
		if err != nil {
			return nil, err
		}

		return strconv.FormatFloat(value2, 'f', decimalSeparator, 64), nil
	})
}

func getRequiredStake(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		blockHeight, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		value := config_stake.GetRequiredStake(blockHeight)
		return strconv.FormatUint(value, 10), nil
	})
}

func getRewardAt(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		blockHeight, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		return strconv.FormatUint(config_reward.GetRewardAt(blockHeight), 10), nil
	})
}

func getNetworkSelectedSeeds(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		value := config.NETWORK_SELECTED_SEEDS
		return webassembly_utils.ConvertJSONBytes(value)
	})
}

func getNetworkDelegatorNodes(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		value := config.NETWORK_SELECTED_DELEGATOR_NODES
		return webassembly_utils.ConvertJSONBytes(value)
	})
}
