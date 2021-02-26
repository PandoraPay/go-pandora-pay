package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
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

	key := []byte("blockHash")
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

	key := []byte("blockHash")
	key = append(key, hash[:]...)

	err := bucket.Put(key, blkComplete.Block.Serialize())
	if err != nil {
		return err
	}

	key = []byte("blockHeight" + strconv.FormatUint(blkComplete.Block.Height, 10))
	return bucket.Put(key, hash[:])

}

func (chain *Blockchain) loadBlockHash(bucket *bolt.Bucket, height uint64) (hash crypto.Hash, err error) {
	if height < 0 || height > chain.Height {
		err = errors.New("Height is invalid")
		return
	}

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	hash = *crypto.ConvertHash(bucket.Get(key))
	return
}

func (chain *Blockchain) saveTotalDifficultyExtra(bucket *bolt.Bucket) error {
	key := []byte("totalDifficulty" + strconv.FormatUint(chain.Height, 10))

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, chain.Timestamp)
	buf = buf[:n]

	buf = append(buf, chain.BigTotalDifficulty.Bytes()...)

	return bucket.Put(key, buf)
}

func loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64, err error) {
	key := []byte("totalDifficulty" + strconv.FormatUint(height, 10))

	buf := bucket.Get(key)
	if buf == nil {
		err = errors.New("Couldn't ready difficulty from DB")
		return
	}

	timestamp, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	difficulty = new(big.Int).SetBytes(buf)

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
