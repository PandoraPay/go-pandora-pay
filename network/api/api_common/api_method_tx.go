package api_common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITransactionRequest struct {
	Height     uint64                  `json:"height,omitempty"`
	Hash       helpers.HexBytes        `json:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APITransactionReply struct {
	Tx           *transaction.Transaction `json:"tx,omitempty"`
	TxSerialized helpers.HexBytes         `json:"serialized,omitempty"`
	Mempool      bool                     `json:"mempool,omitempty"`
	Info         *info.TxInfo             `json:"info,omitempty"`
}

func (api *APICommon) openLoadTx(args *APITransactionRequest, reply *APITransactionReply) error {
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

		reply.Tx = &transaction.Transaction{}
		if err = reply.Tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return err
		}

		if config.SEED_WALLET_NODES_INFO {
			if data = reader.Get("txInfo_ByHash" + hashStr); data == nil {
				return errors.New("TxInfo was not found")
			}
			reply.Info = &info.TxInfo{}
			if err = json.Unmarshal(data, reply.Info); err != nil {
				return err
			}
		}

		return
	})
}

func (api *APICommon) Tx(r *http.Request, args *APITransactionRequest, reply *APITransactionReply) (err error) {

	if len(args.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(args.Hash))
		if txMempool != nil {
			reply.Mempool = true
			reply.Tx = txMempool.Tx
		} else {
			err = api.openLoadTx(args, reply)
		}
	} else {
		err = api.openLoadTx(args, reply)
	}

	if err != nil || reply.Tx == nil {
		return err
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.TxSerialized = reply.Tx.Bloom.Serialized
		reply.Tx = nil
	}

	return
}

func (api *APICommon) GetTx_http(values url.Values) (interface{}, error) {
	args := &APITransactionRequest{0, nil, api_types.RETURN_JSON}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APITransactionReply{}
	return reply, api.Tx(nil, args, reply)
}

func (api *APICommon) GetTx_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APITransactionRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APITransactionReply{}
	return reply, api.Tx(nil, args, reply)
}
