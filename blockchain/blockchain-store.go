package blockchain

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (chain *Blockchain) LoadBlock(bucket *bolt.Bucket, hash []byte) (blk *block.Block) {
	blockData := bucket.Get(append([]byte("blockHash"), hash...))
	if blockData == nil {
		return
	}
	blk = &block.Block{}
	blk.Deserialize(helpers.NewBufferReader(blockData))
	return
}

func (chain *Blockchain) deleteUnusedBlocksComplete(bucket *bolt.Bucket, blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	accs.DeleteTransitionalChangesFromStore(blockHeightStr)
	toks.DeleteTransitionalChangesFromStore(blockHeightStr)

	bucket.Delete([]byte("blockHeight" + blockHeightStr))
	bucket.Delete([]byte("blockTxs" + blockHeightStr))
}

func (chain *Blockchain) removeBlockComplete(bucket *bolt.Bucket, blockHeight uint64, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) {

	blockHeightStr := strconv.FormatUint(blockHeight, 10)
	accs.ReadTransitionalChangesFromStore(blockHeightStr)
	toks.ReadTransitionalChangesFromStore(blockHeightStr)

	hash := bucket.Get([]byte("blockHeight" + blockHeightStr))
	bucket.Delete(append([]byte("blockHash"), hash...))

	data := bucket.Get([]byte("blockTxs" + blockHeightStr))
	txHashes := make([][]byte, 0) //32 byte

	if err := json.Unmarshal(data, &txHashes); err != nil {
		panic(err)
	}
	for _, txHash := range txHashes {
		removedTxHashes[string(txHash)] = txHash
	}

}

func (chain *Blockchain) saveBlockComplete(bucket *bolt.Bucket, blkComplete *block.BlockComplete, hash []byte, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) [][]byte {

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	accs.WriteTransitionalChangesToStore(blockHeightStr)
	toks.WriteTransitionalChangesToStore(blockHeightStr)

	bucket.Put(append([]byte("blockHash"), hash...), blkComplete.Block.Serialize())
	bucket.Put([]byte("blockHeight"+blockHeightStr), hash)

	newTxHashes := make([][]byte, 0)

	txHashes := make([][]byte, 0)
	for _, tx := range blkComplete.Txs {
		txHash := tx.ComputeHash()
		txHashes = append(txHashes, txHash)

		//let's check to see if the tx block is already stored, if yes, we will skip it
		if removedTxHashes[string(txHash)] == nil {
			bucket.Put(append([]byte("tx"), txHash...), tx.Serialize())
			newTxHashes = append(newTxHashes, txHash)
		}
	}

	marshal, err := json.Marshal(txHashes)
	if err != nil {
		panic(err)
	}
	bucket.Put([]byte("blockTxs"+blockHeightStr), marshal)

	return newTxHashes
}

func (chain *Blockchain) LoadBlockHash(bucket *bolt.Bucket, height uint64) []byte {
	if height < 0 {
		panic("Height is invalid")
	}

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	return bucket.Get(key)
}

//chain must be locked before
func (chain *Blockchain) saveTotalDifficultyExtra(bucket *bolt.Bucket) {
	key := []byte("totalDifficulty" + strconv.FormatUint(chain.Height, 10))

	writer := helpers.NewBufferWriter()
	writer.WriteUvarint(chain.Timestamp)

	bytes := chain.BigTotalDifficulty.Bytes()
	writer.WriteUvarint(uint64(len(bytes)))
	writer.Write(bytes)

	bucket.Put(key, writer.Bytes())
}

func (chain *Blockchain) loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64) {
	if height < 0 {
		panic("height is invalid")
	}
	key := []byte("totalDifficulty" + strconv.FormatUint(height, 10))

	buf := bucket.Get(key)
	if buf == nil {
		panic("Couldn't ready difficulty from DB")
	}

	reader := helpers.NewBufferReader(buf)
	timestamp = reader.ReadUvarint()
	length := reader.ReadUvarint()
	bytes := reader.ReadBytes(int(length))
	difficulty = new(big.Int).SetBytes(bytes)
	return
}

func (chain *Blockchain) saveBlockchain(bucket *bolt.Bucket) {
	marshal, err := json.Marshal(chain)
	if err != nil {
		panic(err)
	}

	bucket.Put([]byte("blockchainInfo"), marshal)
}

func (chain *Blockchain) loadBlockchain() (success bool, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) error {

		chain.Lock()
		defer chain.Unlock()

		reader := tx.Bucket([]byte("Chain"))
		chainData := reader.Get([]byte("blockchainInfo"))
		if chainData == nil {
			return nil
		}

		if err = json.Unmarshal(chainData, &chain); err != nil {
			return err
		}
		success = true

		return nil
	})

	return

}
