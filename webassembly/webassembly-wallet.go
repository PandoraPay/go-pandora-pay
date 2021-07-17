package webassembly

import (
	"encoding/hex"
	"pandora-pay/app"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		app.Wallet.RLock()
		defer app.Wallet.RUnlock()
		return convertJSON(app.Wallet)
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[0].String())
		if err != nil {
			return nil, err
		}
		return convertJSON(addr)
	})
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		adr, err := app.Wallet.AddNewAddress(false)
		if err != nil {
			return nil, err
		}
		return convertJSON(adr)
	})
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		return app.Wallet.RemoveAddress(0, args[0].String())
	})
}

func importWalletPrivateKey(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}
		adr, err := app.Wallet.ImportPrivateKey("", key)
		if err != nil {
			return nil, err
		}
		return convertJSON(adr)
	})
}

func importWalletJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		return true, app.Wallet.ImportWalletJSON([]byte(args[0].String()))
	})
}

func importWalletAddressJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		adr, err := app.Wallet.ImportWalletAddressJSON([]byte(args[0].String()))
		if err != nil {
			return nil, err
		}
		return convertJSON(adr)
	})
}

func checkPasswordWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		return app.Wallet.Encryption.CheckPassword(args[0].String())
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
		if err := app.Wallet.Encryption.RemoveEncryption(); err != nil {
			return nil, err
		}
		return true, nil
	})
}
