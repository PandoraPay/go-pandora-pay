package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITxRawRequest struct {
	Height uint64         `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   helpers.Base64 `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

type APITxRawReply struct {
	Tx []byte `json:"tx" msgpack:"tx"`
}

func (api *APICommon) openLoadTxOnly(args *APITxRawRequest, reply *APITxRawReply) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.loadTxHash(reader, args.Height); err != nil {
				return
			}
		}

		hashStr := string(args.Hash)

		if reply.Tx = reader.Get("tx:" + hashStr); reply.Tx == nil {
			return errors.New("Tx not found")
		}

		return
	})
}

func (api *APICommon) GetTxRaw(r *http.Request, args *APITxRawRequest, reply *APITxRawReply) error {

	if len(args.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(args.Hash))
		if txMempool != nil {
			reply.Tx = txMempool.Tx.Bloom.Serialized
			return nil
		}
	}

	return api.openLoadTxOnly(args, reply)
}
