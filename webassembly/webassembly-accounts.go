package webassembly

import (
	"pandora-pay/addresses"
	"syscall/js"
)

func decodeAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		address, err := addresses.DecodeAddr(args[0].String())
		if err != nil {
			return
		}
		return convertJSON(address)
	})
}
