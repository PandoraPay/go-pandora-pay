package webassembly

import (
	"encoding/hex"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"syscall/js"
)

func computePublicKeyHash(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)
		return helpers.HexBytes(publicKeyHash), nil
	})
}
