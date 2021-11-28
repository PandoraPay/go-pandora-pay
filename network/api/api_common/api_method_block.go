package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIBlockRequest struct {
	Height     uint64                  `json:"height,omitempty"`
	Hash       helpers.HexBytes        `json:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APIBlockWithTxsAnswer struct {
	Block           *block.Block       `json:"block,omitempty"`
	BlockSerialized helpers.HexBytes   `json:"serialized,omitempty"`
	Txs             []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) Block(r *http.Request, args *APIBlockRequest, reply *APIBlockWithTxsAnswer) error {

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.chain.LoadBlockHash(reader, args.Height); err != nil {
				return
			}
		}

		if reply.Block, err = api.ApiStore.loadBlock(reader, args.Hash); err != nil || reply.Block == nil {
			return helpers.ReturnErrorIfNot(err, "Block was not found")
		}

		txHashes := [][]byte{}
		data := reader.Get("blockTxs" + strconv.FormatUint(reply.Block.Height, 10))
		if err = json.Unmarshal(data, &txHashes); err != nil {
			return nil
		}

		reply.Txs = make([]helpers.HexBytes, len(txHashes))
		for i, txHash := range txHashes {
			reply.Txs[i] = txHash
		}

		return
	}); err != nil {
		return err
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.BlockSerialized = helpers.SerializeToBytes(reply.Block)
		reply.Block = nil
	}

	return nil
}

func (api *APICommon) GetBlock_http(values url.Values) (interface{}, error) {
	args := &APIBlockRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIBlockWithTxsAnswer{}
	return reply, api.Block(nil, args, reply)
}

func (api *APICommon) GetBlock_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIBlockRequest{0, nil, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIBlockWithTxsAnswer{}
	return reply, api.Block(nil, args, reply)
}
