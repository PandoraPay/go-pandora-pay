package blockchain

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (chain *Blockchain) LoadBlock(bucket *bolt.Bucket, hash []byte) (blk *block.Block, err error) {
	blockData := bucket.Get(append([]byte("blockHash"), hash...))
	if blockData == nil {
		return nil, errors.New("Block was not found")
	}
	blk = &block.Block{}
	err = blk.Deserialize(helpers.NewBufferReader(blockData))
	return
}

func (chain *Blockchain) deleteUnusedBlocksComplete(bucket *bolt.Bucket, blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	if err = accs.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.DeleteTransitionalChangesFromStore(blockHeightStr); err != nil {
		return
	}

	if err = bucket.Delete([]byte("blockHeight" + blockHeightStr)); err != nil {
		return
	}
	if err = bucket.Delete([]byte("blockTxs" + blockHeightStr)); err != nil {
		return
	}

	return
}

func (chain *Blockchain) removeBlockComplete(bucket *bolt.Bucket, blockHeight uint64, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (err error) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	blockHeightNextStr := strconv.FormatUint(blockHeight, 10)

	if err = accs.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}
	if err = toks.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
		return
	}

	hash := bucket.Get([]byte("blockHeight" + blockHeightStr))
	if err = bucket.Delete(append([]byte("blockHash"), hash...)); err != nil {
		return
	}

	data := bucket.Get([]byte("blockTxs" + blockHeightStr))
	txHashes := [][]byte{} //32 byte

	if err = json.Unmarshal(data, &txHashes); err != nil {
		return
	}

	for _, txHash := range txHashes {
		removedTxHashes[string(txHash)] = txHash
	}
	return
}

func (chain *Blockchain) saveBlockComplete(bucket *bolt.Bucket, blkComplete *block_complete.BlockComplete, hash []byte, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) (newTxHashes [][]byte, err error) {

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	if err = accs.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}
	if err = toks.WriteTransitionalChangesToStore(blockHeightStr); err != nil {
		return
	}

	if err = bucket.Put(append([]byte("blockHash"), hash...), blkComplete.Block.SerializeToBytes()); err != nil {
		return
	}
	if err = bucket.Put([]byte("blockHeight"+blockHeightStr), hash); err != nil {
		return
	}

	newTxHashes = [][]byte{}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txHashes[i] = tx.Bloom.Hash

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[tx.Bloom.HashStr] == nil {
			if err = bucket.Put(append([]byte("tx"), tx.Bloom.Hash...), tx.Bloom.Serialized); err != nil {
				return
			}
			newTxHashes = append(newTxHashes, tx.Bloom.Hash)
		}
	}

	marshal, err := json.Marshal(txHashes)
	if err != nil {
		return
	}

	if err = bucket.Put([]byte("blockTxs"+blockHeightStr), marshal); err != nil {
		return
	}

	return
}

func (chain *Blockchain) LoadBlockHash(bucket *bolt.Bucket, height uint64) ([]byte, error) {
	if height < 0 {
		return nil, errors.New("Height is invalid")
	}

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	hash := bucket.Get(key)
	if hash == nil {
		return nil, errors.New("Block Hash not found")
	}
	return hash, nil
}

func (chain *Blockchain) saveBlockchain() error {
	return store.StoreBlockchain.DB.Update(func(boltTx *bolt.Tx) error {
		writer := boltTx.Bucket([]byte("Chain"))
		chainData := chain.GetChainData()
		return chainData.saveBlockchain(writer)
	})
}

func (chain *Blockchain) loadBlockchain() error {

	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		reader := boltTx.Bucket([]byte("Chain"))

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
