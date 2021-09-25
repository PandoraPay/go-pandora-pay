package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/mempool"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"syscall/js"
)

func storeAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		var err error

		publicKey, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		var acc *account.Account
		if !args[1].IsNull() {
			acc = account.NewAccount(publicKey)
			if err = json.Unmarshal([]byte(args[1].String()), &acc); err != nil {
				return nil, err
			}
		}

		mutex.Lock()
		defer mutex.Unlock()

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(mempool.CONTINUE_PROCESSING_NO_ERROR_RESET)

		if err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			accs := accounts.NewAccounts(writer)
			if acc == nil {
				accs.Delete(string(publicKey))
			} else {
				if err = accs.Update(string(publicKey), acc); err != nil {
					return
				}
			}
			return accs.CommitChanges()

		}); err != nil {
			return nil, err
		}

		return true, nil
	})
}

func storeToken(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		var err error

		var tok *token.Token
		if !args[1].IsNull() {
			tok = &token.Token{}
			if err = json.Unmarshal([]byte(args[1].String()), &tok); err != nil {
				return nil, err
			}
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(mempool.CONTINUE_PROCESSING_NO_ERROR_RESET)
		if err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			toks := tokens.NewTokens(writer)
			if tok == nil {
				toks.DeleteToken(hash)
			} else {
				if err = toks.UpdateToken(hash, tok); err != nil {
					return
				}
			}
			return toks.CommitChanges()
		}); err != nil {
			return nil, err
		}

		return true, nil
	})
}
