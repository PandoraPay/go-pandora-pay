package api_common

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountsByKeysRequest struct {
	Keys           []*api_types.APIAccountBaseRequest `json:"keys,omitempty"`
	Asset          helpers.HexBytes                   `json:"asset,omitempty"`
	IncludeMempool bool                               `json:"includeMempool,omitempty"`
	ReturnType     api_types.APIReturnType            `json:"returnType,omitempty"`
}

type APIAccountsByKeysReply struct {
	Acc           []*account.Account           `json:"acc,omitempty"`
	AccSerialized []helpers.HexBytes           `json:"accSerialized,omitempty"`
	Reg           []*registration.Registration `json:"registration,omitempty"`
	RegSerialized []helpers.HexBytes           `json:"registrationSerialized,omitempty"`
}

func (api *APICommon) AccountsByKeys(r *http.Request, args *APIAccountsByKeysRequest, reply *APIAccountsByKeysReply) (err error) {

	publicKeys := make([][]byte, len(args.Keys))

	for i, key := range args.Keys {
		if publicKeys[i], err = key.GetPublicKey(); err != nil {
			return
		}
	}

	if len(publicKeys) > 512*2 {
		return fmt.Errorf("Too many indexes to process: limit %d, found %d", 512*2, len(publicKeys))
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accsCollection := accounts.NewAccountsCollection(reader)
		regs := registrations.NewRegistrations(reader)

		accs, err := accsCollection.GetMap(args.Asset)
		if err != nil {
			return
		}

		reply.Acc = make([]*account.Account, len(publicKeys))
		reply.Reg = make([]*registration.Registration, len(publicKeys))

		for i := 0; i < len(publicKeys); i++ {
			if reply.Acc[i], err = accs.GetAccount(publicKeys[i]); err != nil {
				return
			}
			if reply.Reg[i], err = regs.GetRegistration(publicKeys[i]); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return
	}

	if args.IncludeMempool {
		balancesInit := make([][]byte, len(publicKeys))
		for i, acc := range reply.Acc {
			if acc != nil {
				balancesInit[i] = helpers.SerializeToBytes(acc.Balance)
			}
		}
		if balancesInit, err = api.mempool.GetZetherBalanceMultiple(publicKeys, balancesInit, args.Asset); err != nil {
			return
		}
		for i, acc := range reply.Acc {
			if balancesInit[i] != nil {
				if err = acc.Balance.Deserialize(helpers.NewBufferReader(balancesInit[i])); err != nil {
					return
				}
			}
		}
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.AccSerialized = make([]helpers.HexBytes, len(reply.Acc))
		for i, acc := range reply.Acc {
			if acc != nil {
				reply.AccSerialized[i] = helpers.SerializeToBytes(acc)
			}
		}
		reply.Acc = nil

		reply.RegSerialized = make([]helpers.HexBytes, len(reply.Reg))
		for i, reg := range reply.Reg {
			if reg != nil {
				reply.RegSerialized[i] = helpers.SerializeToBytes(reg)
			}
		}
		reply.Reg = nil
	}
	return
}

func (api *APICommon) GetAccountsByKeys_http(values url.Values) (interface{}, error) {
	args := &APIAccountsByKeysRequest{nil, nil, false, api_types.RETURN_JSON}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountsByKeysReply{}
	return reply, api.AccountsByKeys(nil, args, reply)
}

func (api *APICommon) GetAccountsByKeys_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountsByKeysRequest{nil, nil, false, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountsByKeysReply{}
	return reply, api.AccountsByKeys(nil, args, reply)
}
