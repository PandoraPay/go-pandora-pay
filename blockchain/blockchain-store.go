package blockchain

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
	"sync/atomic"
	"unsafe"
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
	txHashes := [][]byte{} //32 byte

	if err := json.Unmarshal(data, &txHashes); err != nil {
		panic(err)
	}
	for _, txHash := range txHashes {
		removedTxHashes[string(txHash)] = txHash
	}

}

func (chain *Blockchain) saveBlockComplete(bucket *bolt.Bucket, blkComplete *block_complete.BlockComplete, hash []byte, removedTxHashes map[string][]byte, accs *accounts.Accounts, toks *tokens.Tokens) [][]byte {

	blockHeightStr := strconv.FormatUint(blkComplete.Block.Height, 10)
	accs.WriteTransitionalChangesToStore(blockHeightStr)
	toks.WriteTransitionalChangesToStore(blockHeightStr)

	bucket.Put(append([]byte("blockHash"), hash...), blkComplete.Block.Serialize())
	bucket.Put([]byte("blockHeight"+blockHeightStr), hash)

	newTxHashes := [][]byte{}

	txHashes := make([][]byte, len(blkComplete.Txs))
	for i, tx := range blkComplete.Txs {
		txHash := tx.ComputeHash()
		txHashes[i] = txHash

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

func (chain *Blockchain) loadBlockchain() (success bool, err error) {

	err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {

		chain.Lock()
		defer chain.Unlock()

		reader := boltTx.Bucket([]byte("Chain"))
		chainInfoData := reader.Get([]byte("blockchainInfo"))
		if chainInfoData == nil {
			return nil
		}

		chainData := BlockchainData{}

		if err = json.Unmarshal(chainInfoData, &chainData); err != nil {
			return err
		}
		success = true

		atomic.StorePointer(&chain.ChainData, unsafe.Pointer(&chainData))

		return nil
	})

	return

}
