package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common/api_types"
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

		var apiAcc *api_types.APIAccount
		if args[1].Type() == js.TypeString {
			apiAcc = &api_types.APIAccount{}
			if err = json.Unmarshal([]byte(args[1].String()), apiAcc); err != nil {
				return nil, err
			}
		}

		mutex.Lock()
		defer mutex.Unlock()

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(mempool.CONTINUE_PROCESSING_NO_ERROR_RESET)

		if err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			accsCollection := accounts.NewAccountsCollection(writer)
			plainAccs := plain_accounts.NewPlainAccounts(writer)
			regs := registrations.NewRegistrations(writer)

			var accs *accounts.Accounts

			var tokensList [][]byte
			if tokensList, err = accsCollection.GetAccountTokens(publicKey); err != nil {
				return
			}

			for _, token := range tokensList {
				if accs, err = accsCollection.GetMap(token); err != nil {
					return
				}
				accs.Delete(string(publicKey))
			}

			plainAccs.Delete(string(publicKey))
			regs.Delete(string(publicKey))

			if apiAcc != nil {

				for i, token := range apiAcc.Tokens {
					if accs, err = accsCollection.GetMap(token); err != nil {
						return
					}
					apiAcc.Accs[i].PublicKey = publicKey
					apiAcc.Accs[i].Token = token
					if err = accs.Update(string(publicKey), apiAcc.Accs[i]); err != nil {
						return
					}
				}

				if apiAcc.PlainAcc != nil {
					apiAcc.PlainAcc.PublicKey = publicKey
					if err = plainAccs.Update(string(publicKey), apiAcc.PlainAcc); err != nil {
						return
					}
				}

				if apiAcc.Reg != nil {
					apiAcc.Reg.PublicKey = publicKey
					if err = regs.Update(string(publicKey), apiAcc.Reg); err != nil {
						return
					}
				}

			}
			if err = accsCollection.CommitChanges(); err != nil {
				return
			}
			if err = regs.CommitChanges(); err != nil {
				return
			}
			if err = plainAccs.CommitChanges(); err != nil {
				return
			}

			return nil
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
