package main

import (
	"context"
	"pandora-pay/builds/builds_data"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/cryptography/crypto/balance_decryptor"
	"strconv"
	"syscall/js"
	"time"
)

func initializeBalanceDecryptor(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		parameters := &builds_data.WalletInitializeBalanceDecryptorReq{}
		if err := webassembly_utils.UnmarshalBytes(args[0], parameters); err != nil {
			return nil, err
		}

		balance_decryptor.BalanceDecryptor.SetTableSize(parameters.TableSize, ctx, func(status string) {
			args[1].Invoke(status)
		})

		return true, nil
	})
}

func decryptBalance(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		parameters := &builds_data.WalletDecryptBalanceReq{}
		if err := webassembly_utils.UnmarshalBytes(args[0], parameters); err != nil {
			return nil, err
		}

		var value uint64
		var finalErr error
		done := false

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			defer cancel()

			time.Sleep(time.Millisecond * 10)

			value, finalErr = AddressBalanceDecryptor.DecryptBalance("wallet", parameters.PublicKey, parameters.PrivateKey, parameters.Balance, parameters.Asset, true, parameters.PreviousValue, true, ctx, func(status string) {
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
