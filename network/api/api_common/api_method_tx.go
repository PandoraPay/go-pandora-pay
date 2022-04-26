package api_common

import (
	"encoding/binary"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITxRequest struct {
	Height     uint64                  `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash       helpers.Base64          `json:"hash,omitempty" msgpack:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APITxReply struct {
	Tx            *transaction.Transaction `json:"tx,omitempty" msgpack:"tx,omitempty"`
	TxSerialized  []byte                   `json:"serialized,omitempty" msgpack:"serialized,omitempty"`
	Mempool       bool                     `json:"mempool,omitempty" msgpack:"mempool,omitempty"`
	Info          *info.TxInfo             `json:"info,omitempty" msgpack:"info,omitempty"`
	Confirmations uint64                   `json:"confirmations,omitempty" msgpack:"confirmations,omitempty"`
}

func (api *APICommon) openLoadTx(args *APITxRequest, reply *APITxReply) error {
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

		if args.ReturnType == api_types.RETURN_SERIALIZED {
			reply.TxSerialized = data
		} else {
			reply.Tx = &transaction.Transaction{}
			if err = reply.Tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
				return err
			}
		}

		if data = reader.Get("txBlock:" + hashStr); data != nil {
			var blockHeight, chainHeight uint64
			if blockHeight, _ = binary.Uvarint(data); err != nil {
				return err
			}
			if chainHeight, _ = binary.Uvarint(reader.Get("chainHeight")); err != nil {
				return err
			}
			reply.Confirmations = chainHeight - blockHeight
		}

		if config.SEED_WALLET_NODES_INFO {
			if data = reader.Get("txInfo_ByHash" + hashStr); data == nil {
				return errors.New("TxInfo was not found")
			}
			reply.Info = &info.TxInfo{}
			if err = msgpack.Unmarshal(data, reply.Info); err != nil {
				return err
			}
		}

		return
	})
}

func (api *APICommon) GetTx(r *http.Request, args *APITxRequest, reply *APITxReply) error {

	if len(args.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(args.Hash))
		if txMempool != nil {
			reply.Mempool = true
			reply.Tx = txMempool.Tx
			if args.ReturnType == api_types.RETURN_SERIALIZED {
				reply.TxSerialized = reply.Tx.Bloom.Serialized
				reply.Tx = nil
			}
			return nil
		}
	}

	return api.openLoadTx(args, reply)
}
