package api_common

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountRequest struct {
	api_types.APIAccountBaseRequest
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APIAccount struct {
	Accs               []*account.Account                                      `json:"accounts,omitempty"`
	AccsSerialized     []helpers.HexBytes                                      `json:"accountsSerialized,omitempty"`
	AccsExtra          []*api_types.APISubscriptionNotificationAccountExtra    `json:"accountsExtra,omitempty"`
	PlainAcc           *plain_account.PlainAccount                             `json:"plainAccount,omitempty"`
	PlainAccSerialized helpers.HexBytes                                        `json:"plainAccountSerialized,omitempty"`
	PlainAccExtra      *api_types.APISubscriptionNotificationPlainAccExtra     `json:"plainAccountExtra,omitempty"`
	Reg                *registration.Registration                              `json:"registration,omitempty"`
	RegSerialized      helpers.HexBytes                                        `json:"registrationSerialized,omitempty"`
	RegExtra           *api_types.APISubscriptionNotificationRegistrationExtra `json:"registrationExtra,omitempty"`
}

func (api *APICommon) Account(r *http.Request, args *APIAccountRequest, reply *APIAccount) (err error) {

	publicKey, err := args.GetPublicKey()
	if err != nil {
		return
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		accsCollection := accounts.NewAccountsCollection(reader)
		plainAccs := plain_accounts.NewPlainAccounts(reader)
		regs := registrations.NewRegistrations(reader)

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

		if reply.PlainAcc, err = plainAccs.GetPlainAccount(publicKey, chainHeight); err != nil {
			return
		}
		if reply.PlainAcc != nil {
			reply.PlainAccExtra = &api_types.APISubscriptionNotificationPlainAccExtra{
				reply.PlainAcc.Index,
			}
		}

		if reply.Reg, err = regs.GetRegistration(publicKey); err != nil {
			return
		}
		if reply.Reg != nil {
			reply.RegExtra = &api_types.APISubscriptionNotificationRegistrationExtra{
				reply.Reg.Index,
			}
		}

		return
	}); err != nil {
		return err
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {

		reply.AccsSerialized = make([]helpers.HexBytes, len(reply.Accs))
		for i, acc := range reply.Accs {
			reply.AccsSerialized[i] = helpers.SerializeToBytes(acc)
		}
		reply.Accs = nil

		if reply.PlainAcc != nil {
			reply.PlainAccSerialized = helpers.SerializeToBytes(reply.PlainAcc)
			reply.PlainAcc = nil
		}
		if reply.Reg != nil {
			reply.RegSerialized = helpers.SerializeToBytes(reply.Reg)
			reply.Reg = nil
		}

	}

	return
}

func (api *APICommon) GetAccount_http(values url.Values) (interface{}, error) {
	args := &APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.RETURN_JSON}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAccount{}
	return reply, api.Account(nil, args, reply)
}

func (api *APICommon) GetAccount_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAccount{}
	return reply, api.Account(nil, args, reply)
}
