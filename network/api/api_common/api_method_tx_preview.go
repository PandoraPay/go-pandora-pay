package api_common

import (
	"net/http"
	"pandora-pay/blockchain/info"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITransactionPreviewRequest struct {
	Height uint64         `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   helpers.Base64 `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

type APITransactionPreviewReply struct {
	TxPreview *info.TxPreview `json:"txPreview,omitempty" msgpack:"txPreview,omitempty"`
	Mempool   bool            `json:"mempool,omitempty" msgpack:"mempool,omitempty"`
	Info      *info.TxInfo    `json:"info,omitempty" msgpack:"info,omitempty"`
}

func (apiStore *APIStore) openLoadTxPreview(args *APITransactionPreviewRequest, reply *APITransactionPreviewReply) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = apiStore.loadTxHash(reader, args.Height); err != nil {
				return
			}
		}

		reply.TxPreview = &info.TxPreview{}
		if err = apiStore.loadTxPreview(reader, args.Hash, reply.TxPreview); err != nil {
			return
		}
		reply.Info = &info.TxInfo{}
		return apiStore.loadTxInfo(reader, args.Hash, reply.Info)
	})
}

func (api *APICommon) GetTxPreview(r *http.Request, args *APITransactionPreviewRequest, reply *APITransactionPreviewReply) (err error) {

	if args.Hash != nil && len(args.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(args.Hash))
		if txMempool != nil {
			reply.Mempool = true
			if reply.TxPreview, err = info.CreateTxPreviewFromTx(txMempool.Tx); err != nil {
				return
			}
		} else {
			err = api.ApiStore.openLoadTxPreview(args, reply)
		}
	} else {
		err = api.ApiStore.openLoadTxPreview(args, reply)
	}

	return
}
