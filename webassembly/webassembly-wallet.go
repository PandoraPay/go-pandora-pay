package webassembly

import (
	"encoding/hex"
	"pandora-pay/config/globals"
	"pandora-pay/wallet"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		wallet := globals.Data["wallet"].(*wallet.Wallet)
		wallet.RLock()
		defer wallet.RUnlock()
		return convertJSON(wallet)
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		addr, err := globals.Data["wallet"].(*wallet.Wallet).GetWalletAddressByEncodedAddress(args[0].String())
		if err != nil {
			return
		}
		return convertJSON(addr)
	})
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		adr, err := globals.Data["wallet"].(*wallet.Wallet).AddNewAddress()
		if err != nil {
			return
		}
		return convertJSON(adr)
	})
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (interface{}, error) {
		return globals.Data["wallet"].(*wallet.Wallet).RemoveAddress(0, args[0].String())
	})
}

func importWalletPrivateKey(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		key, err := hex.DecodeString(args[0].String())
		if err != nil {
			return
		}
		adr, err := globals.Data["wallet"].(*wallet.Wallet).ImportPrivateKey("", key)
		if err != nil {
			return
		}
		return convertJSON(adr)
	})
}

func importWalletJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		return true, globals.Data["wallet"].(*wallet.Wallet).ImportWalletJSON([]byte(args[0].String()))
	})
}

func importWalletAddressJSON(this js.Value, args []js.Value) interface{} {
	return promiseFunction(this, args, func(js.Value, []js.Value) (out interface{}, err error) {
		adr, err := globals.Data["wallet"].(*wallet.Wallet).ImportWalletAddressJSON([]byte(args[0].String()))
		if err != nil {
			return
		}
		return convertJSON(adr)
	})
}
