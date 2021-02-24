package blockchain

import (
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

	adr := []byte("block")
	adr = append(adr, hash[:]...)

	blockData := bucket.Get(adr)
	if blockData == nil {
		return
	}

	blk = &block.Block{}
	_, err = blk.Deserialize(blockData)

	return
}
