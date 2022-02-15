package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/cryptography"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITransactionRawRequest struct {
	Height uint64 `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   []byte `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

func (api *APICommon) openLoadTxOnly(args *APITransactionRawRequest, reply *[]byte) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.loadTxHash(reader, args.Height); err != nil {
				return
			}
		}

		hashStr := string(args.Hash)
		var data []byte

		if data = reader.Get("tx:" + hashStr); data == nil {
			return errors.New("Tx not found")
		}

		*reply = data

		return
	})
}

func (api *APICommon) TxRaw(r *http.Request, args *APITransactionRawRequest, reply *[]byte) error {

	if len(args.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(args.Hash))
		if txMempool != nil {
			*reply = txMempool.Tx.Bloom.Serialized
			return nil
		}
	}

	return api.openLoadTxOnly(args, reply)
}

func (api *APICommon) GetTxRaw_http(values url.Values) (interface{}, error) {
	args := &APITransactionRawRequest{0, nil}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	var reply []byte
	return reply, api.TxRaw(nil, args, &reply)
}

func (api *APICommon) GetTxRaw_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APITransactionRawRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var reply []byte
	return reply, api.TxRaw(nil, args, &reply)
}
