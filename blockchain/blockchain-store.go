package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	block_info "pandora-pay/blockchain/block-info"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func (chain *Blockchain) LoadBlockInfo(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (blk *block_info.BlockInfo, err error) {
	blockData := reader.Get(append([]byte("blockInfo_ByHash"), hash...))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk = &block_info.BlockInfo{}
	err = json.Unmarshal(blockData, blk)
	return
}

func (chain *Blockchain) LoadBlock(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (blk *block.Block, err error) {
	blockData := reader.Get(append([]byte("block_ByHash"), hash...))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk = &block.Block{BlockHeader: &block.BlockHeader{}}
	err = blk.Deserialize(helpers.NewBufferReader(blockData))
	return
}

func (chain *Blockchain) deleteUnusedBlocksComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	if err = accs.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}

	if err = writer.Delete([]byte("blockHash_ByHeight" + blockHeightStr)); err != nil {
		return
	}
	if err = writer.Delete([]byte("blockTxs" + blockHeightStr)); err != nil {
		return
	}

	return
}

func (chain *Blockchain) removeBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	blockHeightNextStr := strconv.FormatUint(blockHeight, 10)

	if err = accs.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}
	if err = toks.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}

	hash := writer.Get([]byte("blockHash_ByHeight" + blockHeightStr))
	if err = writer.Delete(append([]byte("block_ByHash"), hash...)); err != nil {
		return
	}
	if err = writer.Delete(append([]byte("blockInfo_ByHash"), hash...)); err != nil {
		return
	}

	data := writer.Get([]byte("blockTxs" + blockHeightStr))
	txHashes := [][]byte{} //32 byte

	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	for _, txHash := range txHashes {
		removedTxHashes[string(txHash)] = txHash
	}
	return
}

func (chain *Blockchain) saveBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, transactionsCount uint64, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (newTxHashes [][]byte, err error) {

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	if err = accs.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}

	if err = writer.Put(append([]byte("block_ByHash"), blkComplete.Block.Bloom.Hash...), blkComplete.Block.SerializeToBytes()); err != nil {
		return
	}

	blockInfoMarshal, err := json.Marshal(&block_info.BlockInfo{
		Hash:       blkComplete.Block.Bloom.Hash,
		KernelHash: blkComplete.Block.Bloom.KernelHash,
		Timestamp:  blkComplete.Block.Timestamp,
		Size:       blkComplete.BloomBlkComplete.Size,
		TXs:        uint64(len(blkComplete.Txs)),
		Forger:     blkComplete.Block.Forger,
	})

	if err != nil {
		return
	}
	if err = writer.Put(append([]byte("blockInfo_ByHash"), blkComplete.Block.Bloom.Hash...), blockInfoMarshal); err != nil {
		return
	}

	if err = writer.Put([]byte("blockHash_ByHeight"+blockHeightStr), blkComplete.Block.Bloom.Hash); err != nil {
		return
	}

	newTxHashes = [][]byte{}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {

		txHashes[i] = tx.Bloom.Hash

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[tx.Bloom.HashStr] == nil {
			if err = writer.Put(append([]byte("tx"), tx.Bloom.Hash...), tx.SerializeToBytesBloomed()); err != nil {
				return
			}
			newTxHashes = append(newTxHashes, tx.Bloom.Hash)
		}

		indexStr := strconv.FormatUint(transactionsCount+uint64(i), 10)
		if err = writer.Put([]byte("txHash_ByHeight"+indexStr), tx.Bloom.Hash); err != nil {
			return
		}
	}

	marshal, err := json.Marshal(txHashes)
	if err != nil {
		return
	}

	if err = writer.Put([]byte("blockTxs"+blockHeightStr), marshal); err != nil {
		return
	}

	return
}

func (chain *Blockchain) LoadBlockHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}

	hash := reader.Get([]byte("blockHash_ByHeight" + strconv.FormatUint(height, 10)))
	if hash == nil {
		return nil, errors.New("Block Hash not found")
	}
	return hash, nil
}

func (chain *Blockchain) LoadTxHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}

	hash := reader.Get([]byte("txHash_ByHeight" + strconv.FormatUint(height, 10)))
	if hash == nil {
		return nil, errors.New("Tx Hash not found")
	}
	return hash, nil
}

func (chain *Blockchain) saveBlockchain() error {
	return store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) error {
		chainData := chain.GetChainData()
		return chainData.saveBlockchain(writer)
	})
}

func (chain *Blockchain) loadBlockchain() error {

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainInfoData := reader.Get([]byte("blockchainInfo"))
		if chainInfoData == nil {
			return errors.New("Chain not found")
		}

		chainData := &BlockchainData{}

		if err = json.Unmarshal(chainInfoData, chainData); err != nil {
			return err
		}
		chain.ChainData.Store(chainData)

		return
	})

}
