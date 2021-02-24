package blockchain

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/block"
	"pandora-pay/crypto"
	"pandora-pay/store"
)

func (chain *Blockchain) LoadBlockFromHash(hash crypto.Hash) (blk *block.Block, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Chain"))
		if reader == nil {
			return nil
		}

		blk, err = LoadBlock(reader, hash)

		return err
	})

	return

}

func LoadBlock(bucket *bolt.Bucket, hash crypto.Hash) (blk *block.Block, err error) {

	key := []byte("block")
	key = append(key, hash[:]...)

	blockData := bucket.Get(key)
	if blockData == nil {
		return
	}

	blk = &block.Block{}
	_, err = blk.Deserialize(blockData)

	return
}

func SaveBlock(bucket *bolt.Bucket, blkComplete *block.BlockComplete) error {

	hash := blkComplete.Block.ComputeHash()

	key := []byte("block")
	key = append(key, hash[:]...)

	return bucket.Put(key, blkComplete.Block.Serialize())

}

func SaveBlockchain(bucket *bolt.Bucket, chain *Blockchain) error {

	marshal, err := json.Marshal(chain)
	if err != nil {
		return errors.New("Error marshaling chain")
	}

	err = bucket.Put([]byte("chain"), marshal)
	return err

}
