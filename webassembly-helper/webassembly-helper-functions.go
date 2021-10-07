package main

import (
	"fmt"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

func helloPandoraHelper(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		fmt.Println("HelloPandoraHelper works!")
		return true, nil
	})
}
