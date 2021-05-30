package webassembly

import (
	"encoding/hex"
	"pandora-pay/config/globals"
	"pandora-pay/wallet"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			wallet := globals.Data["wallet"].(*wallet.Wallet)
			wallet.RLock()
			out := convertJSON(wallet)
			wallet.RUnlock()
			args2[0].Invoke(out)
		}()
		return nil
	}))
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			adr, err := globals.Data["wallet"].(*wallet.Wallet).GetWalletAddressByEncodedAddress(args[1].String())
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(convertJSON(adr))
		}()
		return nil
	}))
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			adr, err := globals.Data["wallet"].(*wallet.Wallet).AddNewAddress()
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(convertJSON(adr))
		}()
		return nil
	}))
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			adr, err := globals.Data["wallet"].(*wallet.Wallet).RemoveAddress(0, args[0].String())
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(convertJSON(adr))
		}()
		return nil
	}))
}

func importWalletPrivateKey(this js.Value, args []js.Value) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {

			key, err := hex.DecodeString(args[0].String())
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}

			adr, err := globals.Data["wallet"].(*wallet.Wallet).ImportPrivateKey("", key)
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(convertJSON(adr))
		}()
		return nil
	}))
}
