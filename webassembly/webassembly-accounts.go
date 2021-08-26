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
		var key, registration, paymentId []byte
		var amount uint64

		if key, err = hex.DecodeString(args[0].String()); err != nil {
			return nil, err
		}

		if registration, err = hex.DecodeString(args[1].String()); err != nil {
			return nil, err
		}

		amount = uint64(args[2].Int())

		if paymentId, err = hex.DecodeString(args[3].String()); err != nil {
			return nil, err
		}

		addr, err := addresses.CreateAddr(key, registration, amount, paymentId)
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
