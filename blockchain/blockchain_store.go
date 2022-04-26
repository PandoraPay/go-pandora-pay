package blockchain

import (
	"encoding/binary"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

func (chain *Blockchain) OpenExistsTx(hash []byte) (exists bool, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		exists = reader.Exists("txHash:" + string(hash)) //optimized
		return nil
	})
	return
}

func (chain *Blockchain) OpenExistsBlock(hash []byte) (exists bool, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		exists = reader.Exists("blockHeight_ByHash" + string(hash)) //optimized
		return nil
	})
	return
}

func (chain *Blockchain) OpenExistsAsset(hash []byte) (exists bool, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		exists = reader.Exists("assets:exists:" + string(hash)) //optimized
		return nil
	})
	return
}

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

func (chain *Blockchain) deleteUnusedBlocksComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, dataStorage *data_storage.DataStorage) error {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)

	if err := dataStorage.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return err
	}

	writer.Delete("blockHash_ByHeight" + blockHeightStr)
	writer.Delete("blockKernelHash_ByHeight" + blockHeightStr)
	writer.Delete("blockTxs" + blockHeightStr)

	return nil
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

	writer.Delete("block_ByHash" + string(hash))
	writer.Delete("blockHeight_ByHash" + string(hash))
	writer.Delete("blockHash_ByHeight" + string(blockHeightStr))
	writer.Delete("blockKernelHash_ByHeight" + string(blockHeightStr))

	data := writer.Get("blockTxs" + blockHeightStr)
	txHashes := [][]byte{} //32 byte

	if err := msgpack.Unmarshal(data, &txHashes); err != nil {
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
	//it will commit the changes
	if err := dataStorage.CommitChanges(); err != nil {
		return allTransactionsChanges, err
	}

	writer.Put("block_ByHash"+string(blkComplete.Block.Bloom.Hash), helpers.SerializeToBytes(blkComplete.Block))
	writer.Put("blockHash_ByHeight"+blockHeightStr, blkComplete.Block.Bloom.Hash)
	writer.Put("blockKernelHash_ByHeight"+blockHeightStr, blkComplete.Block.Bloom.KernelHash)
	writer.Put("blockHeight_ByHash"+string(blkComplete.Block.Bloom.Hash), []byte(blockHeightStr))

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txHashes[i] = tx.Bloom.Hash
	}
	marshal, err := msgpack.Marshal(txHashes)
	if err != nil {
		return allTransactionsChanges, err
	}
	writer.Put("blockTxs"+blockHeightStr, marshal)

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
			writer.Put("tx:"+tx.Bloom.HashStr, tx.Bloom.Serialized)
			writer.Put("txHash:"+tx.Bloom.HashStr, []byte{1})

			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, blkComplete.Block.Height)
			writer.Put("txBlock:"+tx.Bloom.HashStr, buf[:n])
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

	if config.SEED_WALLET_NODES_INFO {
		if err = saveAssetsInfo(dataStorage.Asts); err != nil {
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

		if err = msgpack.Unmarshal(chainInfoData, chainData); err != nil {
			return err
		}
		chain.ChainData.Store(chainData)

		return
	})

}
