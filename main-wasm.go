// +build wasm

package main

import (
	"pandora-pay/gui"
	"syscall/js"
)

func HelloPandora(js.Value, []js.Value) interface{} {
	gui.GUI.Info("HelloPandora works!")
	return nil
}

func additionalMain() {
	js.Global().Set("HelloPandora", js.FuncOf(HelloPandoraWASM))
}
