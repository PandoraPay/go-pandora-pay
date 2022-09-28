package main

import (
	"encoding/base64"
	"fmt"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/helpers"
	"pandora-pay/helpers/identicon"
	"pandora-pay/start"
	"strconv"
	"syscall/js"
)

func helloPandora(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		fmt.Println("HelloPandora works!")
		return true, nil
	})
}

func startLibrary(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		err := start.StartMainNow()
		return true, err
	})
}

func getIdenticon(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		publicKey, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		identicon, err := identicon.GenerateToBytes(publicKey, args[1].Int(), args[2].Int())
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertBytes(identicon), nil
	})
}

func randomUint64(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		return helpers.RandomUint64(), nil
	})
}

func randomUint64N(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		n, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}
		return strconv.FormatUint(helpers.RandomUint64()%n, 10), nil
	})
}

func shuffleArray(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		n, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		array := helpers.ShuffleArray(int(n))
		return webassembly_utils.ConvertJSONBytes(array)
	})
}

func shuffleArray_for_Zether(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		n, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		array := helpers.ShuffleArray_for_Zether(int(n))
		return webassembly_utils.ConvertJSONBytes(array)
	})
}
