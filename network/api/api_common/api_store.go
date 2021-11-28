package api_common

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIStore struct {
	chain *blockchain.Blockchain
}

func (apiStore *APIStore) loadTxInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte, reply *info.TxInfo) error {
	data := reader.Get("txInfo_ByHash" + string(hash))
	if data == nil {
		return errors.New("TxInfo was not found")
	}
	return json.Unmarshal(data, reply)
}

func (apiStore *APIStore) loadTxPreview(reader store_db_interface.StoreDBTransactionInterface, hash []byte, reply *info.TxPreview) error {
	data := reader.Get("txPreview_ByHash" + string(hash))
	if data == nil {
		return errors.New("TxPreview was not found")
	}
	return json.Unmarshal(data, reply)
}

func (apiStore *APIStore) loadAssetHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}
	return reader.Get("assets:list:" + strconv.FormatUint(height, 10)), nil
}

func (apiStore *APIStore) loadTxHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}
	return reader.Get("txHash_ByHeight" + strconv.FormatUint(height, 10)), nil
}

func (chain *APIStore) loadBlock(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (*block.Block, error) {
	blockData := reader.Get("block_ByHash" + string(hash))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk := block.CreateEmptyBlock()
	return blk, blk.Deserialize(helpers.NewBufferReader(blockData))
}

func NewAPIStore(chain *blockchain.Blockchain) *APIStore {
	return &APIStore{
		chain: chain,
	}
}
