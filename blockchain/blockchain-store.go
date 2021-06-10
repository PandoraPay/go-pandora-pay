package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func (chain *Blockchain) OpenLoadBlockHash(blockHeight uint64) (hash []byte, errfinal error) {
	errfinal = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
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

func (chain *Blockchain) deleteUnusedBlocksComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)

	if err = accs.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
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

func (chain *Blockchain) removeBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blockHeight uint64, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	blockHeightNextStr := strconv.FormatUint(blockHeight, 10)

	if err = accs.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}
	if err = toks.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}

	hash := writer.Get("blockHash_ByHeight" + blockHeightStr)
	if hash == nil {
		return errors.New("Invalid Hash")
	}

	if err = writer.Delete("block_ByHash" + string(hash)); err != nil {
		return
	}

	data := writer.Get("blockTxs" + blockHeightStr)
	txHashes := [][]byte{} //32 byte

	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	for _, txHash := range txHashes {
		removedTxHashes[string(txHash)] = txHash
	}

	if config.SEED_WALLET_NODES_INFO {
		if err = removeBlockCompleteInfo(writer, hash, txHashes); err != nil {
			return
		}
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

	if err = writer.Put("block_ByHash"+string(blkComplete.Block.Bloom.Hash), blkComplete.Block.SerializeToBytes()); err != nil {
		return
	}

	if err = writer.Put("blockHash_ByHeight"+blockHeightStr, blkComplete.Block.Bloom.Hash); err != nil {
		return
	}

	newTxHashes = [][]byte{}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {

		txHashes[i] = tx.Bloom.Hash

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[tx.Bloom.HashStr] == nil {
			if err = writer.Put("tx"+string(tx.Bloom.Hash), tx.SerializeToBytesBloomed()); err != nil {
				return
			}
			newTxHashes = append(newTxHashes, tx.Bloom.Hash)
		}

	}

	if config.SEED_WALLET_NODES_INFO {
		if err = saveBlockCompleteInfo(writer, blkComplete, transactionsCount); err != nil {
			return
		}
	}

	var marshal []byte
	if marshal, err = json.Marshal(txHashes); err != nil {
		return
	}

	if err = writer.Put("blockTxs"+blockHeightStr, marshal); err != nil {
		return
	}

	return
}

func (chain *Blockchain) saveBlockchainHashmaps(accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	accs.Rollback()
	toks.Rollback()

	if err = accs.WriteToStore(); err != nil {
		return
	}
	if err = toks.WriteToStore(); err != nil {
		return
	}

	if err = accs.CloneCommitted(); err != nil {
		return
	}
	if err = toks.CloneCommitted(); err != nil {
		return
	}

	if config.SEED_WALLET_NODES_INFO {
		if err = saveTokensInfo(toks); err != nil {
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
