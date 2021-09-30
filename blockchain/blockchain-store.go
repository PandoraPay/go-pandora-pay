package blockchain

import (
	"encoding/json"
	"errors"
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func (chain *Blockchain) OpenLoadBlockHash(blockHeight uint64) (hash []byte, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		hash, err = chain.LoadBlockHash(reader, blockHeight)
		return
	})
	return
}

func (chain *Blockchain) LoadBlockHash(reader store_db_interface.StoreDBTransactionInterface, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}

	hash := reader.Get("blockHash_ByHeight" + strconv.FormatUint(height, 10))
	if hash == nil {
		return nil, errors.New("Block Hash not found")
	}
	return hash, nil
}

func (chain *Blockchain) deleteUnusedBlocksComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)

	if err = dataStorage.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}

	if err = writer.Delete("blockHash_ByHeight" + blockHeightStr); err != nil {
		return
	}
	if err = writer.Delete("blockTxs" + blockHeightStr); err != nil {
		return
	}

	return
}

func (chain *Blockchain) removeBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, removedTxHashes map[string][]byte, allTransactionsChanges []*blockchain_types.BlockchainTransactionUpdate, dataStorage *data_storage.DataStorage) (allTransactionsChanges2 []*blockchain_types.BlockchainTransactionUpdate, err error) {

	allTransactionsChanges2 = allTransactionsChanges
	allTransactionsChangesFinal := allTransactionsChanges

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	blockHeightNextStr := strconv.FormatUint(blockHeight, 10)

	if err = dataStorage.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}

	hash := writer.Get("blockHash_ByHeight" + blockHeightStr)
	if hash == nil {
		return allTransactionsChanges, errors.New("Invalid Hash")
	}

	if err = writer.Delete("block_ByHash" + string(hash)); err != nil {
		return
	}
	if err = writer.Delete("blockHeight_ByHash" + string(hash)); err != nil {
		return
	}

	data := writer.Get("blockTxs" + blockHeightStr)
	txHashes := [][]byte{} //32 byte

	if err := json.Unmarshal(data, &txHashes); err != nil {
		return allTransactionsChanges, err
	}

	localTransactionChanges := make([]*blockchain_types.BlockchainTransactionUpdate, len(txHashes))
	for i, txHash := range txHashes {

		txChange := &blockchain_types.BlockchainTransactionUpdate{
			TxHash:    txHash,
			TxHashStr: string(txHash),
			Inserted:  false,
		}

		allTransactionsChangesFinal = append(allTransactionsChangesFinal, txChange)
		localTransactionChanges[i] = txChange

		removedTxHashes[txChange.TxHashStr] = txHash
	}

	if config.SEED_WALLET_NODES_INFO {
		if err = removeBlockCompleteInfo(writer, hash, txHashes, localTransactionChanges); err != nil {
			return
		}
	}

	return allTransactionsChangesFinal, nil
}

func (chain *Blockchain) saveBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, transactionsCount uint64, removedTxHashes map[string][]byte, allTransactionsChanges []*blockchain_types.BlockchainTransactionUpdate, dataStorage *data_storage.DataStorage) ([]*blockchain_types.BlockchainTransactionUpdate, error) {

	allTransactionsChanges2 := allTransactionsChanges

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	if err := dataStorage.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return allTransactionsChanges, err
	}

	if err := writer.Put("block_ByHash"+string(blkComplete.Block.Bloom.Hash), blkComplete.Block.SerializeToBytes()); err != nil {
		return allTransactionsChanges, err
	}

	if err := writer.Put("blockHash_ByHeight"+blockHeightStr, blkComplete.Block.Bloom.Hash); err != nil {
		return allTransactionsChanges, err
	}

	if err := writer.Put("blockHeight_ByHash"+string(blkComplete.Block.Bloom.Hash), []byte(blockHeightStr)); err != nil {
		return allTransactionsChanges, err
	}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txHashes[i] = tx.Bloom.Hash
	}
	marshal, err := json.Marshal(txHashes)
	if err != nil {
		return allTransactionsChanges, err
	}
	if err = writer.Put("blockTxs"+blockHeightStr, marshal); err != nil {
		return allTransactionsChanges, err
	}

	localTransactionChanges := make([]*blockchain_types.BlockchainTransactionUpdate, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {

		txChange := &blockchain_types.BlockchainTransactionUpdate{
			TxHash:         tx.Bloom.Hash,
			TxHashStr:      tx.Bloom.HashStr,
			Tx:             tx,
			Inserted:       true,
			BlockHeight:    blkComplete.Block.Height,
			BlockTimestamp: blkComplete.Block.Timestamp,
			Height:         transactionsCount + uint64(i),
		}

		allTransactionsChanges2 = append(allTransactionsChanges2, txChange)
		localTransactionChanges[i] = txChange

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[tx.Bloom.HashStr] == nil {
			if err := writer.Put("tx"+tx.Bloom.HashStr, tx.Bloom.Serialized); err != nil {
				return allTransactionsChanges, err
			}
		} else {
			delete(removedTxHashes, tx.Bloom.HashStr)
		}

	}

	if config.SEED_WALLET_NODES_INFO {
		if err := saveBlockCompleteInfo(writer, blkComplete, transactionsCount, localTransactionChanges); err != nil {
			return allTransactionsChanges, err
		}
	}

	return allTransactionsChanges2, nil
}

func (chain *Blockchain) saveBlockchainHashmaps(dataStorage *data_storage.DataStorage) (err error) {

	dataStorage.Rollback()
	if err = dataStorage.CloneCommitted(); err != nil {
		return
	}

	if config.SEED_WALLET_NODES_INFO {
		if err = saveTokensInfo(dataStorage.Toks); err != nil {
			return
		}
	}

	return
}

func (chain *Blockchain) saveBlockchain() error {
	return store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) error {
		chainData := chain.GetChainData()
		return chainData.saveBlockchain(writer)
	})
}

func (chain *Blockchain) loadBlockchain() error {

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainInfoData := reader.Get("blockchainInfo")
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
