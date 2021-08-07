package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/mempool"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"syscall/js"
)

func storeAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		var err error

		var acc *account.Account
		if !args[1].IsNull() {
			acc = &account.Account{}
			if err = json.Unmarshal([]byte(args[1].String()), &acc); err != nil {
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

			accs := accounts.NewAccounts(writer)
			if acc == nil {
				accs.Delete(string(hash))
			} else {
				if err = accs.UpdateAccount(hash, acc); err != nil {
					return
				}
			}
			accs.CommitChanges()
			return accs.WriteToStore()
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
			toks.CommitChanges()
			return toks.WriteToStore()
		}); err != nil {
			return nil, err
		}

		return true, nil
	})
}
