package api_common

import (
	"net/http"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITransactionInfoRequest struct {
	Height uint64         `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   helpers.Base64 `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

func (api *APICommon) GetTxInfo(r *http.Request, args *APITransactionInfoRequest, reply *info.TxInfo) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.loadTxHash(reader, args.Height); err != nil {
				return
			}
		}

		return api.ApiStore.loadTxInfo(reader, args.Hash, reply)
	})
}
