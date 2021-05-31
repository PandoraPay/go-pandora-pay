package webassembly

import (
	"encoding/base64"
	"pandora-pay/gui"
	"pandora-pay/helpers/identicon"
	"syscall/js"
)

func helloPandora(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		gui.GUI.Info("HelloPandora works!")
		return true, nil
	})
}

func start(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		startMainCallback()
		return true, nil
	})
}

func getIdenticon(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		identicon, err := identicon.GenerateToBytes([]byte(args[0].String()), args[1].Int(), args[2].Int())
		if err != nil {
			return
		}
		return "data:image/png;base64," + base64.StdEncoding.EncodeToString(identicon), nil
	})
}
