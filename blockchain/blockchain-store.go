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

		blk, err = loadBlock(reader, hash)

		return err
	})

	return

}

func loadBlock(bucket *bolt.Bucket, hash crypto.Hash) (blk *block.Block, err error) {

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

func saveBlock(bucket *bolt.Bucket, blkComplete *block.BlockComplete) error {

	hash := blkComplete.Block.ComputeHash()

	key := []byte("block")
	key = append(key, hash[:]...)

	return bucket.Put(key, blkComplete.Block.Serialize())

}

func saveBlockchain(bucket *bolt.Bucket, chain *Blockchain) error {

	marshal, err := json.Marshal(chain)
	if err != nil {
		return errors.New("Error marshaling chain")
	}

	err = bucket.Put([]byte("blockchainInfo"), marshal)
	return err

}

func loadBlockchain() (success bool, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Chain"))
		if reader == nil {
			return nil
		}

		chainData := reader.Get([]byte("blockchainInfo"))
		if chainData == nil {
			return nil
		}

		err = json.Unmarshal(chainData, &Chain)
		if err != nil {
			return err
		}
		success = true

		return nil
	})

	return

}
