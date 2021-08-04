package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"syscall/js"
)

func storeAccount(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		var err error

		acc := &account.Account{}

		if !args[1].IsNull() {
			if err = json.Unmarshal([]byte(args[1].String()), &acc); err != nil {
				return nil, err
			}
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		if err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			accs := accounts.NewAccounts(writer)
			if err = accs.UpdateAccount(hash, acc); err != nil {
				return
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

		tok := &token.Token{}
		if !args[1].IsNull() {
			if err = json.Unmarshal([]byte(args[1].String()), &tok); err != nil {
				return nil, err
			}
		}

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		if err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			toks := tokens.NewTokens(writer)
			if err = toks.UpdateToken(hash, tok); err != nil {
				return
			}
			toks.CommitChanges()
			return toks.WriteToStore()
		}); err != nil {
			return nil, err
		}

		return true, nil
	})
}
