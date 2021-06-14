package webassembly

import (
	"encoding/hex"
	"pandora-pay/addresses"
	"syscall/js"
)

func decodeAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		address, err := addresses.DecodeAddr(args[0].String())
		if err != nil {
			return nil, err
		}
		return convertJSON(address)
	})
}

func generateAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		var err error
		var key, paymentId []byte
		var amount uint64

		if key, err = hex.DecodeString(args[0].String()); err != nil {
			return nil, err
		}

		if len(args) >= 2 {
			amount = uint64(args[1].Int())
		}

		if len(args) >= 3 {
			if paymentId, err = hex.DecodeString(args[2].String()); err != nil {
				return nil, err
			}
		}

		addr, err := addresses.CreateAddr(key, amount, paymentId)
		if err != nil {
			return nil, err
		}

		json, err := convertJSON(addr)
		if err != nil {
			return nil, err
		}

		return []interface{}{
			json,
			addr.EncodeAddr(),
		}, nil

	})
}
