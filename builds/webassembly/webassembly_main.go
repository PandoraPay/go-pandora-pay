package main

import (
	"pandora-pay/start"
	"syscall/js"
)

func main() {
	start.InitMain(func() {
		Initialize()
		js.Global().Call("WASMLoaded")
	})
}
