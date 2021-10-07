package webassembly

import (
	"encoding/hex"
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

func decodeAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		address, err := addresses.DecodeAddr(args[0].String())
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(address)
	})
}

func generateAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := struct {
			PublicKey    helpers.HexBytes `json:"publicKey"`
			Registration helpers.HexBytes `json:"registration"`
			PaymentId    helpers.HexBytes `json:"paymentId"`
			Amount       uint64           `json:"amount"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], &parameters); err != nil {
			return nil, err
		}

		addr, err := addresses.CreateAddr(parameters.PublicKey, parameters.Registration, parameters.Amount, parameters.PaymentId)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			addr,
			addr.EncodeAddr(),
		})

	})
}

func generateNewAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		priv := addresses.GenerateNewPrivateKey()
		addr, err := priv.GenerateAddress(true, 0, nil)

		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes([]interface{}{
			hex.EncodeToString(priv.Key),
			addr.EncodeAddr(),
			hex.EncodeToString(addr.PublicKey),
		})
	})
}
