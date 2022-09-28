package main

import (
	"encoding/base64"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/helpers"
	"syscall/js"
)

func getWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		app.Wallet.Lock.RLock()
		defer app.Wallet.Lock.RUnlock()

		data, err := helpers.GetJSONDataExcept(app.Wallet, "mnemonic", "seed")
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(data), nil
	})
}

func exportWalletJSON(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		app.Wallet.Lock.RLock()
		defer app.Wallet.Lock.RUnlock()

		data, err := helpers.GetJSONDataExcept(app.Wallet)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertBytes(data), nil
	})
}

func getWalletMnemonic(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		app.Wallet.Lock.RLock()
		defer app.Wallet.Lock.RUnlock()
		return app.Wallet.Mnemonic, nil
	})
}

func getWalletSeed(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		app.Wallet.Lock.RLock()
		defer app.Wallet.Lock.RUnlock()
		return base64.StdEncoding.EncodeToString(app.Wallet.Seed), nil
	})
}

func createNewWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		return nil, app.Wallet.CreateEmptyWallet()
	})
}

func importMnemonic(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		return true, app.Wallet.ImportMnemonic(args[1].String())
	})
}

func getWalletAddressSecretKey(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByPublicKeyHashString(args[1].String(), true)
		if err != nil {
			return nil, err
		}

		return base64.StdEncoding.EncodeToString(addr.SecretKey), nil
	})
}

func getWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		adr, err := app.Wallet.GetWalletAddressByPublicKeyHashString(args[1].String(), true)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(adr)
	})
}

func addNewWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		adr, err := app.Wallet.AddNewAddress(false, args[1].String(), true)
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(adr)
	})
}

func removeWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		publicKey, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		return app.Wallet.RemoveAddressByPublicKeyHash(publicKey, true)
	})
}

func renameWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		publicKey, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		return app.Wallet.RenameAddressByPublicKeyHash(publicKey, args[2].String(), true)
	})
}

func importWalletSecretKey(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}

		key, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}
		adr, err := app.Wallet.ImportSecretKey(args[2].String(), key)

		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(adr)
	})
}

func importWalletJSON(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		return true, app.Wallet.ImportWalletJSON([]byte(args[1].String()))
	})
}

func importWalletAddressJSON(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return nil, err
		}
		adr, err := app.Wallet.ImportWalletAddressJSON([]byte(args[1].String()))
		if err != nil {
			return nil, err
		}
		return webassembly_utils.ConvertJSONBytes(adr)
	})
}

func checkPasswordWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}
		return true, nil
	})
}

func encryptWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Encrypt(args[0].String(), args[1].Int()); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func decryptWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Decrypt(args[0].String()); err != nil {
			return nil, err
		}
		return true, nil
	})
}

func removeEncryptionWallet(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
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
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.Logout(); err != nil {
			return nil, err
		}
		return true, nil
	})
}

//signing not encrypting
func signMessageWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}

		message, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[2].String(), true)
		if err != nil {
			return nil, err
		}

		out, err := addr.SignMessage(message)
		if err != nil {
			return nil, err
		}

		return base64.StdEncoding.EncodeToString(out), nil
	})
}

func decryptMessageWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {
		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}

		data, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[2].String(), true)
		if err != nil {
			return nil, err
		}

		out, err := addr.DecryptMessage(data)
		if err != nil {
			return nil, err
		}

		return base64.StdEncoding.EncodeToString(out), nil
	})
}

func deriveSharedStakedWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}

		addr, err := app.Wallet.GetWalletAddressByEncodedAddress(args[1].String(), true)
		if err != nil {
			return nil, err
		}

		sharedStaked, err := addr.DeriveSharedStaked(uint32(args[2].Int()))
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(sharedStaked)

	})
}

func getPrivateKeysWalletAddress(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if err := app.Wallet.Encryption.CheckPassword(args[0].String(), false); err != nil {
			return false, err
		}

		parameters := &struct {
			PublicKeyHash []byte `json:"publicKeyHash"`
			Asset         []byte `json:"asset"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[1], parameters); err != nil {
			return nil, err
		}

		privateKey := app.Wallet.GetPrivateKeys(parameters.PublicKeyHash, parameters.Asset)

		return webassembly_utils.ConvertJSONBytes(struct {
			PrivateKey []byte `json:"privateKey"`
		}{privateKey})

	})
}

func decryptTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		publicKeyHash, err := base64.StdEncoding.DecodeString(args[1].String())
		if err != nil {
			return nil, err
		}

		data := webassembly_utils.GetBytes(args[0])

		tx := &transaction.Transaction{}
		if err := tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}

		decrypted, err := app.Wallet.DecryptTx(tx, publicKeyHash)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(decrypted)
	})
}
