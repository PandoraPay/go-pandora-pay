package webassembly

import (
	"encoding/base64"
	"pandora-pay/gui"
	"pandora-pay/helpers/identicon"
	"syscall/js"
)

func helloPandora(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			gui.GUI.Info("HelloPandora works!")
			args2[0].Invoke(true)
		}()
		return nil
	}))
}

func start(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			startMainCallback()
			args2[0].Invoke(true)
		}()
		return nil
	}))
}

func getIdenticon(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			out, err := identicon.GenerateToBytes([]byte(args[0].String()), args[1].Int(), args[2].Int())
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke("data:image/png;base64," + base64.StdEncoding.EncodeToString(out))
		}()
		return nil
	}))
}
