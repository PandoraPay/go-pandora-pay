package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/helpers"
	"pandora-pay/network/api_code/api_code_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIBlockRequest struct {
	Height     uint64                       `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash       helpers.Base64               `json:"hash,omitempty" msgpack:"hash,omitempty"`
	ReturnType api_code_types.APIReturnType `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APIBlockReply struct {
	Block           *block.Block `json:"block,omitempty" msgpack:"block,omitempty"`
	BlockSerialized []byte       `json:"serialized,omitempty" msgpack:"serialized,omitempty"`
	Txs             [][]byte     `json:"txs,omitempty" msgpack:"txs,omitempty"`
}

func (api *APICommon) GetBlock(r *http.Request, args *APIBlockRequest, reply *APIBlockReply) error {

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
		if err = msgpack.Unmarshal(data, &txHashes); err != nil {
			return nil
		}

		reply.Txs = make([][]byte, len(txHashes))
		for i, txHash := range txHashes {
			reply.Txs[i] = txHash
		}

		return
	}); err != nil {
		return err
	}

	if args.ReturnType == api_code_types.RETURN_SERIALIZED {
		reply.BlockSerialized = helpers.SerializeToBytes(reply.Block)
		reply.Block = nil
	}

	return nil
}
