package api_common

import (
	"fmt"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountsKeysByIndexRequest struct {
	Indexes         []uint64       `json:"indexes" msgpack:"indexes"`
	Asset           helpers.Base64 `json:"asset" msgpack:"asset"`
	EncodeAddresses bool           `json:"encodeAddresses" msgpack:"encodeAddresses"`
}

type APIAccountsKeysByIndexReply struct {
	PublicKeys [][]byte `json:"publicKeys,omitempty" msgpack:"publicKeys,omitempty"`
	Addresses  []string `json:"addresses,omitempty" msgpack:"addresses,omitempty"`
}

func (api *APICommon) GetAccountsKeysByIndex(r *http.Request, args *APIAccountsKeysByIndexRequest, reply *APIAccountsKeysByIndexReply) (err error) {

	if len(args.Indexes) > 512*2 {
		return fmt.Errorf("Too many indexes to process: limit %d, found %d", 512*2, len(args.Indexes))
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accs, err := accounts.NewAccountsCollection(reader).GetMap(args.Asset)
		if err != nil {
			return
		}

		reply.PublicKeys = make([][]byte, len(args.Indexes))
		for i := 0; i < len(args.Indexes); i++ {
			if reply.PublicKeys[i], err = accs.GetKeyByIndex(args.Indexes[i]); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return
	}

	if args.EncodeAddresses {
		reply.Addresses = make([]string, len(reply.PublicKeys))
		for i, publicKey := range reply.PublicKeys {
			var addr *addresses.Address
			if addr, err = addresses.CreateAddr(publicKey, nil, nil, 0, nil); err != nil {
				return
			}
			reply.Addresses[i] = addr.EncodeAddr()
		}
		reply.PublicKeys = nil
	}
	return
}
