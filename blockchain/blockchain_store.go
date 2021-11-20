package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/config"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

func (chain *Blockchain) OpenExistsTx(hash []byte) (exists bool, errFinal error) {
	errFinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		exists = reader.Exists("txs:exists:" + string(hash))
		return
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

	hash := reader.Get("blocks:list:" + strconv.FormatUint(height, 10))
	if hash == nil {
		return nil, errors.New("Block Hash not found")
	}
	return hash, nil
}

func (chain *Blockchain) saveBlockchainHashmaps(writer store_db_interface.StoreDBTransactionInterface, dataStorage *data_storage.DataStorage) (err error) {

	dataStorage.Rollback()
	if err = dataStorage.CloneCommitted(); err != nil {
		return
	}

	if config.SEED_WALLET_NODES_INFO {
		if err = saveExtra(writer, dataStorage); err != nil {
			return
		}
	}

	return
}

func (chain *Blockchain) saveBlock(blkComplete *block_complete.BlockComplete, dataStorage *data_storage.DataStorage) (err error) {

	txs := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txs[i] = tx.Bloom.Hash
	}

	if _, err = dataStorage.Blocks.CreateNewBlock(blkComplete.Bloom.Hash, blkComplete, blkComplete.Block.SerializeManualToBytes(), txs); err != nil {
		return
	}

	for _, tx := range blkComplete.Txs {
		if _, err = dataStorage.Txs.CreateNewTx(tx.Bloom.Hash, tx, tx.Bloom.Serialized); err != nil {
			return
		}
	}

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	if err = dataStorage.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}
	//it will commit the changes
	if err = dataStorage.CommitChanges(); err != nil {
		return
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
