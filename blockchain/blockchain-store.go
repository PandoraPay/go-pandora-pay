package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func (chain *Blockchain) LoadBlock(reader store_db_interface.StoreDBTransactionInterface, hash []byte) (blk *block.Block, err error) {
	blockData := reader.Get(append([]byte("blockHash"), hash...))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk = &block.Block{}
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

	if err = writer.Delete([]byte("blockHeight" + blockHeightStr)); err != nil {
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

	hash := writer.Get([]byte("blockHeight" + blockHeightStr))
	if err = writer.Delete(append([]byte("blockHash"), hash...)); err != nil {
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

func (chain *Blockchain) saveBlockComplete(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, hash []byte, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (newTxHashes [][]byte, err error) {

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	if err = accs.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}

	if err = writer.Put(append([]byte("blockHash"), hash...), blkComplete.Block.SerializeToBytes()); err != nil {
		return
	}
	if err = writer.Put([]byte("blockHeight"+blockHeightStr), hash); err != nil {
		return
	}

	newTxHashes = [][]byte{}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txHashes[i] = tx.Bloom.Hash

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[tx.Bloom.HashStr] == nil {
			if err = writer.Put(append([]byte("tx"), tx.Bloom.Hash...), tx.Bloom.Serialized); err != nil {
				return
			}
			newTxHashes = append(newTxHashes, tx.Bloom.Hash)
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

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	hash := reader.Get(key)
	if hash == nil {
		return nil, errors.New("Block Hash not found")
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
