package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountsCountRequest struct {
	Asset helpers.HexBytes `json:"asset" msgpack:"asset"`
}

type APIAccountsCountReply struct {
	Count uint64 `json:"count" msgpack:"count"`
}

func (api *APICommon) AccountsCount(r *http.Request, args *APIAccountsCountRequest, reply *APIAccountsCountReply) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		accs, err := accounts.NewAccountsCollection(reader).GetMap(args.Asset)
		if err != nil {
			return
		}

		reply.Count = accs.Count
		return
	})
}

func (api *APICommon) GetAccountsCount_http(values url.Values) (interface{}, error) {
	args := &APIAccountsCountRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAccountsCountReply{}
	return reply, api.AccountsCount(nil, args, reply)
}

func (api *APICommon) GetAccountsCount_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountsCountRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountsCountReply{}
	return reply, api.AccountsCount(nil, args, reply)
}
