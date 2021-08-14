package webassembly

import (
	"encoding/hex"
	"pandora-pay/cryptography"
	"syscall/js"
)

func computePublicKeyHash(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)
		return hex.EncodeToString(publicKeyHash), nil
	})
}
