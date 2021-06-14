package webassembly

import (
	"encoding/base64"
	"encoding/hex"
	"pandora-pay/gui"
	"pandora-pay/helpers/identicon"
	"syscall/js"
)

func helloPandora(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		gui.GUI.Info("HelloPandora works!")
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

		publicKeyHash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		identicon, err := identicon.GenerateToBytes(publicKeyHash, args[1].Int(), args[2].Int())
		if err != nil {
			return nil, err
		}

		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(identicon), nil
	})
}
