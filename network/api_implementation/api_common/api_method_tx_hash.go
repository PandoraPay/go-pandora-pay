package api_common

import (
	"net/http"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITxHashRequest struct {
	Height uint64 `json:"height" msgpack:"height"`
}

type APITxHashReply struct {
	Hash []byte `json:"height" msgpack:"height"`
}

func (api *APICommon) GetTxHash(r *http.Request, args *APITxHashRequest, reply *APITxHashReply) (err error) {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		reply.Hash, err = api.ApiStore.loadTxHash(reader, args.Height)
		return
	})
}
