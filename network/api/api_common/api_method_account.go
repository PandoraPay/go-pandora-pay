package api_common

import (
	"net/http"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountRequest struct {
	api_types.APIAccountBaseRequest
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"  msgpack:"returnType,omitempty" `
}

type APIAccountReply struct {
	Accs               []*account.Account                                   `json:"accounts,omitempty" msgpack:"accounts,omitempty"`
	AccsSerialized     [][]byte                                             `json:"accountsSerialized,omitempty" msgpack:"accountsSerialized,omitempty"`
	AccsExtra          []*api_types.APISubscriptionNotificationAccountExtra `json:"accountsExtra,omitempty" msgpack:"accountsExtra,omitempty"`
	PlainAcc           *plain_account.PlainAccount                          `json:"plainAccount,omitempty" msgpack:"plainAccount,omitempty"`
	PlainAccSerialized []byte                                               `json:"plainAccountSerialized,omitempty" msgpack:"plainAccountSerialized,omitempty"`
	PlainAccExtra      *api_types.APISubscriptionNotificationPlainAccExtra  `json:"plainAccountExtra,omitempty" msgpack:"plainAccountExtra,omitempty"`
}

func (api *APICommon) GetAccount(r *http.Request, args *APIAccountRequest, reply *APIAccountReply) (err error) {

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accsCollection := accounts.NewAccountsCollection(reader)
		plainAccs := plain_accounts.NewPlainAccounts(reader)

		assetsList, err := accsCollection.GetAccountAssets(publicKey)
		if err != nil {
			return
		}

		reply.Accs = make([]*account.Account, len(assetsList))
		reply.AccsExtra = make([]*api_types.APISubscriptionNotificationAccountExtra, len(assetsList))

		for i, assetId := range assetsList {

			var accs *accounts.Accounts
			if accs, err = accsCollection.GetMap(assetId); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(publicKey); err != nil {
				return
			}

			reply.Accs[i] = acc
			if acc != nil {
				reply.AccsExtra[i] = &api_types.APISubscriptionNotificationAccountExtra{
					assetId,
					acc.Index,
				}
			}
		}

		if reply.PlainAcc, err = plainAccs.GetPlainAccount(publicKey); err != nil {
			return
		}
		if reply.PlainAcc != nil {
			reply.PlainAccExtra = &api_types.APISubscriptionNotificationPlainAccExtra{
				reply.PlainAcc.Index,
			}
		}

		return
	}); err != nil {
		return err
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {

		reply.AccsSerialized = make([][]byte, len(reply.Accs))
		for i, acc := range reply.Accs {
			reply.AccsSerialized[i] = helpers.SerializeToBytes(acc)
		}
		reply.Accs = nil

		if reply.PlainAcc != nil {
			reply.PlainAccSerialized = helpers.SerializeToBytes(reply.PlainAcc)
			reply.PlainAcc = nil
		}

	}

	return
}
