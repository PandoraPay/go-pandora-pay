package api_common

import (
	"net/http"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountsCountRequest struct {
	Asset helpers.Base64 `json:"asset" msgpack:"asset"`
}

type APIAccountsCountReply struct {
	Count uint64 `json:"count" msgpack:"count"`
}

func (api *APICommon) GetAccountsCount(r *http.Request, args *APIAccountsCountRequest, reply *APIAccountsCountReply) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		accs, err := accounts.NewAccountsCollection(reader).GetMap(args.Asset)
		if err != nil {
			return
		}

		reply.Count = accs.Count
		return
	})
}
