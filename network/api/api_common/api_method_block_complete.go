package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIBlockCompleteRequest struct {
	Height     uint64                  `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash       helpers.HexBytes        `json:"hash,omitempty" msgpack:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APIBlockCompleteReply struct {
	BlockComplete *block_complete.BlockComplete `json:"blockComplete,omitempty" msgpack:"blockComplete,omitempty"`
	Serialized    helpers.HexBytes              `json:"serialized,omitempty" msgpack:"serialized,omitempty"`
}

func (api *APICommon) BlockComplete(r *http.Request, args *APIBlockCompleteRequest, reply *APIBlockCompleteReply) error {

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		if len(args.Hash) == 0 {
			args.Hash, err = api.ApiStore.chain.LoadBlockHash(reader, args.Height)
		}

		reply.BlockComplete = &block_complete.BlockComplete{}

		if reply.BlockComplete.Block, err = api.ApiStore.loadBlock(reader, args.Hash); err != nil || reply.BlockComplete.Block == nil {
			return helpers.ReturnErrorIfNot(err, "Block was not found")
		}

		data := reader.Get("blockTxs" + strconv.FormatUint(reply.BlockComplete.Block.Height, 10))
		if data == nil {
			return errors.New("Strange. blockTxs was not found")
		}

		txHashes := [][]byte{}
		if err = msgpack.Unmarshal(data, &txHashes); err != nil {
			return
		}

		reply.BlockComplete.Txs = make([]*transaction.Transaction, len(txHashes))
		for i, txHash := range txHashes {
			data = reader.Get("tx:" + string(txHash))
			reply.BlockComplete.Txs[i] = &transaction.Transaction{}
			if err = reply.BlockComplete.Txs[i].Deserialize(helpers.NewBufferReader(data)); err != nil {
				return
			}
		}

		return reply.BlockComplete.BloomCompleteBySerialized(reply.BlockComplete.SerializeManualToBytes())
	}); err != nil {
		return err
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.Serialized = reply.BlockComplete.BloomBlkComplete.Serialized
		reply.BlockComplete = nil
	}

	return nil
}

func (api *APICommon) GetBlockComplete_http(values url.Values) (interface{}, error) {
	args := &APIBlockCompleteRequest{0, nil, api_types.RETURN_JSON}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIBlockCompleteReply{}
	return reply, api.BlockComplete(nil, args, reply)
}

func (api *APICommon) GetBlockComplete_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIBlockCompleteRequest{0, nil, api_types.RETURN_SERIALIZED}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIBlockCompleteReply{}
	return reply, api.BlockComplete(nil, args, reply)
}
