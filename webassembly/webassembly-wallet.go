package webassembly

import (
	"encoding/hex"
	"pandora-pay/app"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		app.Wallet.RLock()
		defer app.Wallet.RUnlock()
		return convertJSON(app.Wallet)
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[0].String())
		if err != nil {
			return
		}
		return convertJSON(addr)
	})
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		adr, err := app.Wallet.AddNewAddress()
		if err != nil {
			return
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
	return promiseFunction(func() (out interface{}, err error) {
		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return
		}
		adr, err := app.Wallet.ImportPrivateKey("", key)
		if err != nil {
			return
		}
		return convertJSON(adr)
	})
}

func importWalletJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		return true, app.Wallet.ImportWalletJSON([]byte(args[0].String()))
	})
}

func importWalletAddressJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (out interface{}, err error) {
		adr, err := app.Wallet.ImportWalletAddressJSON([]byte(args[0].String()))
		if err != nil {
			return
		}
		return convertJSON(adr)
	})
}
