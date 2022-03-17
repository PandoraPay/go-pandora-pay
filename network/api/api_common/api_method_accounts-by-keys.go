package api_common

import (
	"fmt"
	"net/http"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/accounts/account/account_balance_homomorphic"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_builder/wizard"
)

type APIAccountsByKeysRequest struct {
	Keys           []*api_types.APIAccountBaseRequest `json:"keys,omitempty" msgpack:"keys,omitempty"`
	Asset          helpers.Base64                     `json:"asset,omitempty" msgpack:"asset,omitempty"`
	IncludeMempool bool                               `json:"includeMempool,omitempty" msgpack:"includeMempool,omitempty"`
	ReturnType     api_types.APIReturnType            `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APIAccountsByKeysReply struct {
	Acc           []*account.Account           `json:"account,omitempty" msgpack:"account,omitempty"`
	AccSerialized [][]byte                     `json:"accountSerialized,omitempty" msgpack:"accountSerialized,omitempty"`
	Reg           []*registration.Registration `json:"registration,omitempty" msgpack:"registration,omitempty"`
	RegSerialized [][]byte                     `json:"registrationSerialized,omitempty" msgpack:"registrationSerialized,omitempty"`
}

func (api *APICommon) GetAccountsByKeys(r *http.Request, args *APIAccountsByKeysRequest, reply *APIAccountsByKeysReply) (err error) {

	publicKeys := make([][]byte, len(args.Keys))
	hasRollovers := make([]bool, len(args.Keys))

	for i, key := range args.Keys {
		if publicKeys[i], err = key.GetPublicKey(true); err != nil {
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
			hasRollovers[i] = reply.Acc[i] != nil && reply.Reg[i].Stakable
		}

		return
	}); err != nil {
		return
	}

	if args.IncludeMempool {
		balancesInit := make([]*crypto.ElGamal, len(publicKeys))
		for i, acc := range reply.Acc {
			if acc != nil {
				balancesInit[i] = acc.Balance.Amount
			}
		}
		if balancesInit, err = wizard.GetZetherBalanceMultiple(publicKeys, balancesInit, args.Asset, hasRollovers, false, api.mempool.Txs.GetTxsOnlyList()); err != nil {
			return
		}
		for i, acc := range reply.Acc {
			if balancesInit[i] != nil {
				acc.Balance = &account_balance_homomorphic.BalanceHomomorphic{nil, balancesInit[i]}
			}
		}
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.AccSerialized = make([][]byte, len(reply.Acc))
		for i, acc := range reply.Acc {
			if acc != nil {
				reply.AccSerialized[i] = helpers.SerializeToBytes(acc)
			}
		}
		reply.Acc = nil

		reply.RegSerialized = make([][]byte, len(reply.Reg))
		for i, reg := range reply.Reg {
			if reg != nil {
				reply.RegSerialized[i] = helpers.SerializeToBytes(reg)
			}
		}
		reply.Reg = nil
	}
	return
}
