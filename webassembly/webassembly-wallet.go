package webassembly

import (
	"encoding/hex"
	"pandora-pay/app"
	"pandora-pay/helpers"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		app.Wallet.RLock()
		defer app.Wallet.RUnlock()

		data, err := helpers.GetJSON(app.Wallet, "mnemonic")
		if err != nil {
			return nil, err
		}
		return string(data), nil
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

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String())
		if err != nil {
			return nil, err
		}
		return hex.EncodeToString(addr.PrivateKey.Key), nil
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String())
		if err != nil {
			return nil, err
		}
		return convertJSON(addr)
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
		return convertJSON(adr)
	})
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		return app.Wallet.RemoveAddress(0, args[1].String())
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
		return convertJSON(adr)
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
		return convertJSON(adr)
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
