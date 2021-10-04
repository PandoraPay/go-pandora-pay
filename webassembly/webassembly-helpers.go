package webassembly

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/helpers"
	"pandora-pay/helpers/identicon"
	"strconv"
	"syscall/js"
)

func helloPandora(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		fmt.Println("HelloPandora works!")
		return true, nil
	})
}

func start(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		startMainCallback()
		return true, nil
	})
}

func getIdenticon(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		identicon, err := identicon.GenerateToBytes(publicKey, args[1].Int(), args[2].Int())
		if err != nil {
			return nil, err
		}
		return convertBytes(identicon)
	})
}

func randomUint64(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		return helpers.RandomUint64(), nil
	})
}

func randomUint64N(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		n, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}
		return helpers.RandomUint64() % n, nil
	})
}
