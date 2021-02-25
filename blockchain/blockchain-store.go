package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/block"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
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

func saveBlock(bucket *bolt.Bucket, blkComplete *block.BlockComplete, hash crypto.Hash) error {

	key := []byte("block")
	key = append(key, hash[:]...)

	return bucket.Put(key, blkComplete.Block.Serialize())

}

func saveTotalDifficulty(bucket *bolt.Bucket, chain *Blockchain) error {
	key := []byte("totalDifficulty" + strconv.Itoa(int(chain.Height)))
	return bucket.Put(key, chain.BigTotalDifficulty.Bytes())
}

func saveTimestamp(bucket *bolt.Bucket, chain *Blockchain) error {
	key := []byte("timestamp" + strconv.Itoa(int(chain.Height)))
	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, chain.Timestamp)
	return bucket.Put(key, buf[:n])
}

func loadTotalDifficulty(bucket *bolt.Bucket, height uint64) *big.Int {
	key := []byte("totalDifficulty" + strconv.Itoa(int(height)))
	buff := bucket.Get(key)
	if buff == nil {
		return nil
	}
	return new(big.Int).SetBytes(buff)
}

func loadTimestamp(bucket *bolt.Bucket, height uint64) (n uint64) {
	key := []byte("timestamp" + strconv.Itoa(int(height)))
	buff := bucket.Get(key)
	n, buff, _ = helpers.DeserializeNumber(buff)
	return
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
