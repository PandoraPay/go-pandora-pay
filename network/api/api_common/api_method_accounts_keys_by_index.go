package api_common

import (
	"encoding/json"
	"fmt"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountsKeysByIndexRequest struct {
	Indexes         []uint64         `json:"indexes"`
	Asset           helpers.HexBytes `json:"asset"`
	EncodeAddresses bool             `json:"encodeAddresses"`
}

type APIAccountsKeysByIndexAnswer struct {
	PublicKeys []helpers.HexBytes `json:"publicKeys,omitempty"`
	Addresses  []string           `json:"addresses,omitempty"`
}

func (api *APICommon) AccountsKeysByIndex(r *http.Request, args *APIAccountsKeysByIndexRequest, reply *APIAccountsKeysByIndexAnswer) (err error) {

	if len(args.Indexes) > 512*2 {
		return fmt.Errorf("Too many indexes to process: limit %d, found %d", 512*2, len(args.Indexes))
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accs, err := accounts.NewAccountsCollection(reader).GetMap(args.Asset)
		if err != nil {
			return
		}

		reply.PublicKeys = make([]helpers.HexBytes, len(args.Indexes))
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

func (api *APICommon) GetAccountsKeysByIndex_http(values url.Values) (interface{}, error) {
	args := &APIAccountsKeysByIndexRequest{nil, nil, true}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountsKeysByIndexAnswer{}
	return reply, api.AccountsKeysByIndex(nil, args, reply)
}

func (api *APICommon) GetAccountsKeysByIndex_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountsKeysByIndexRequest{nil, nil, false}
	if err := json.Unmarshal(values, &args); err != nil {
		return nil, err
	}
	reply := &APIAccountsKeysByIndexAnswer{}
	return reply, api.AccountsKeysByIndex(nil, args, reply)
}
