package main

import (
	"context"
	"pandora-pay/addresses"
	"pandora-pay/app"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/cryptography/crypto/balance-decryptor"
	"pandora-pay/webassembly/webassembly_utils"
	"strconv"
	"syscall/js"
	"time"
)

func initializeBalanceDecryptor(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		balance_decryptor.BalanceDecryptor.SetTableSize(args[0].Int(), ctx, func(status string) {
			args[1].Invoke(status)
		})

		return true, nil
	})
}

func decryptBalance(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := &struct {
			PrivateKey    []byte `json:"privateKey"`
			PreviousValue uint64 `json:"previousValue"`
			Balance       []byte `json:"balance"`
			Asset         []byte `json:"asset"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], parameters); err != nil {
			return nil, err
		}

		privateKey := &addresses.PrivateKey{Key: parameters.PrivateKey}

		balance, err := new(crypto.ElGamal).Deserialize(parameters.Balance)
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

			value, finalErr = app.AddressBalanceDecryptor.DecryptBalanceByPrivateKey(parameters.PrivateKey, balance, parameters.Asset, true, parameters.PreviousValue, true, ctx, func(status string) {
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
					strconv.FormatUint(value, 10),
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
