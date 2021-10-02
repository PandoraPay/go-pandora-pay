package webassembly

import (
	"context"
	"encoding/hex"
	"pandora-pay/app"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"strconv"
	"syscall/js"
	"time"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		app.Wallet.RLock()
		defer app.Wallet.RUnlock()

		data, err := helpers.GetJSON(app.Wallet, "mnemonic")
		if err != nil {
			return nil, err
		}

		return convertBytes(data)
	})
}

func getWalletMnemonic(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		app.Wallet.RLock()
		defer app.Wallet.RUnlock()
		return app.Wallet.Mnemonic, nil
	})
}

func getWalletAddressPrivateKey(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[1].String(), false); err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[0].String())
		if err != nil {
			return nil, err
		}
		return hex.EncodeToString(addr.PrivateKey.Key), nil
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[1].String(), false); err != nil {
			return nil, err
		}

		adr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[0].String())
		if err != nil {
			return nil, err
		}

		return convertJSONBytes(adr)
	})
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		adr, err := app.Wallet.AddNewAddress(false)
		if err != nil {
			return nil, err
		}
		return convertJSONBytes(adr)
	})
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		return app.Wallet.RemoveAddress(args[1].String(), true)
	})
}

func importWalletPrivateKey(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		key, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}
		adr, err := app.Wallet.ImportPrivateKey(args[2].String(), key)
		if err != nil {
			return nil, err
		}
		return convertJSONBytes(adr)
	})
}

func importWalletJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		return true, app.Wallet.ImportWalletJSON([]byte(args[1].String()))
	})
}

func importWalletAddressJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		adr, err := app.Wallet.ImportWalletAddressJSON([]byte(args[1].String()))
		if err != nil {
			return nil, err
		}
		return convertJSONBytes(adr)
	})
}

func checkPasswordWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}
		return true, nil
	})
}

func encryptWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Encrypt(args[0].String(), args[1].Int()); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func decryptWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Decrypt(args[0].String()); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func removeEncryptionWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), true); err != nil {
			return nil, err
		}
		if err := app.Wallet.Encryption.RemoveEncryption(); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func logoutWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Logout(); err != nil {
			return nil, err
		}
		return true, nil
	})
}

//signing not encrypting
func signMessageWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return false, err
		}

		message, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		app.Wallet.RLock()
		defer app.Wallet.RUnlock()

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String())
		if err != nil {
			return nil, err
		}

		out, err := addr.SignMessage(message)
		if err != nil {
			return nil, err
		}

		return hex.EncodeToString(out), nil
	})
}

func decryptMessageWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return false, err
		}

		data, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		app.Wallet.RLock()
		defer app.Wallet.RUnlock()

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String())
		if err != nil {
			return nil, err
		}

		out, err := addr.DecryptMessage(data)
		if err != nil {
			return nil, err
		}

		return hex.EncodeToString(out), nil
	})
}

func deriveDelegatedStakeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return false, err
		}

		nonce, err := strconv.ParseUint(args[0].String(), 10, 64)
		if err != nil {
			return nil, err
		}

		app.Wallet.RLock()
		defer app.Wallet.RUnlock()

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String())
		if err != nil {
			return nil, err
		}

		delegatedStake, err := addr.DeriveDelegatedStake(uint32(nonce))
		if err != nil {
			return nil, err
		}

		return convertJSONBytes(delegatedStake)

	})
}

func decodeBalanceWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[3].String(), false); err != nil {
			return false, err
		}

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		balanceEncoded, err := hex.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		token, err := hex.DecodeString(args[2].String())
		if err != nil {
			return nil, err
		}

		balance, err := new(crypto.ElGamal).Deserialize(balanceEncoded)
		if err != nil {
			return nil, err
		}

		var value uint64
		var finalErr error
		done := false

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			defer cancel()
			value, finalErr = app.Wallet.DecodeBalanceByPublicKey(publicKey, balance, token, true, true, ctx, func(status string) {
				args[4].Invoke(status)
				time.Sleep(1 * time.Millisecond)
			})
			done = true
		}()

		return []interface{}{
			js.FuncOf(func(a js.Value, b []js.Value) interface{} {

				var out interface{}
				if finalErr != nil {
					out = errorConstructor.New(finalErr.Error())
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
