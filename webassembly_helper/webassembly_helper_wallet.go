package main

import (
	"context"
	"pandora-pay/addresses"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
	"time"
)

func initializeBalanceDecoder(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		crypto.BalanceDecoder.SetTableSize(args[0].Int(), ctx, func(status string) {
			args[1].Invoke(status)
		})

		return true, nil
	})
}

func decodeBalance(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := &struct {
			PrivateKey     helpers.HexBytes `json:"privateKey"`
			PreviousValue  uint64           `json:"previousValue"`
			BalanceEncoded helpers.HexBytes `json:"balanceEncoded"`
			Asset          helpers.HexBytes `json:"asset"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], parameters); err != nil {
			return nil, err
		}

		privateKey := &addresses.PrivateKey{Key: parameters.PrivateKey}

		balance, err := new(crypto.ElGamal).Deserialize(parameters.BalanceEncoded)
		if err != nil {
			return nil, err
		}

		var value uint64
		var finalErr error
		done := false

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			defer cancel()

			time.Sleep(time.Millisecond * 10)

			value, finalErr = privateKey.DecodeBalance(balance, parameters.PreviousValue, ctx, func(status string) {
				args[1].Invoke(status)
				time.Sleep(500 * time.Microsecond)
			})

			done = true
		}()

		return []interface{}{
			js.FuncOf(func(a js.Value, b []js.Value) interface{} {

				var out interface{}
				if finalErr != nil {
					out = webassembly_utils.ErrorConstructor.New(finalErr.Error())
				} else {
					out = nil
				}

				return []interface{}{
					done,
					value,
					out,
				}
			}),
			js.FuncOf(func(a js.Value, b []js.Value) interface{} {
				cancel()
				return nil
			}),
		}, nil

	})
}
