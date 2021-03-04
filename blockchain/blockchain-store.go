package blockchain

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (chain *Blockchain) LoadBlockFromHash(hash helpers.Hash) (blk *block.Block, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Chain"))
		if reader == nil {
			return nil
		}

		blk, err = chain.loadBlock(reader, hash)

		return err
	})

	return

}

func (chain *Blockchain) loadBlock(bucket *bolt.Bucket, hash helpers.Hash) (blk *block.Block, err error) {

	key := []byte("blockHash")
	key = append(key, hash[:]...)

	blockData := bucket.Get(key)
	if blockData == nil {
		return
	}

	blk = &block.Block{}

	reader := helpers.NewBufferReader(blockData)
	err = blk.Deserialize(reader)

	return
}

func (chain *Blockchain) saveBlock(bucket *bolt.Bucket, blkComplete *block.BlockComplete, hash helpers.Hash) error {

	key := []byte("blockHash")
	key = append(key, hash[:]...)

	err := bucket.Put(key, blkComplete.Block.Serialize())
	if err != nil {
		return err
	}

	key = []byte("blockHeight" + strconv.FormatUint(blkComplete.Block.Height, 10))
	return bucket.Put(key, hash[:])

}

func (chain *Blockchain) loadBlockHash(bucket *bolt.Bucket, height uint64) (hash helpers.Hash, err error) {
	if height < 0 || height > chain.Height {
		err = errors.New("Height is invalid")
		return
	}

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	hash = *helpers.ConvertHash(bucket.Get(key))
	return
}

func (chain *Blockchain) saveTotalDifficultyExtra(bucket *bolt.Bucket) error {
	key := []byte("totalDifficulty" + strconv.FormatUint(chain.Height, 10))

	writer := helpers.NewBufferWriter()
	writer.WriteUvarint(chain.Timestamp)

	bytes := chain.BigTotalDifficulty.Bytes()
	writer.WriteUvarint(uint64(len(bytes)))
	writer.Write(bytes)

	return bucket.Put(key, writer.Bytes())
}

func (chain *Blockchain) loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64, err error) {
	key := []byte("totalDifficulty" + strconv.FormatUint(height, 10))

	buf := bucket.Get(key)
	if buf == nil {
		err = errors.New("Couldn't ready difficulty from DB")
		return
	}

	reader := helpers.NewBufferReader(buf)
	timestamp, err = reader.ReadUvarint()
	if err != nil {
		return
	}

	var length uint64
	length, err = reader.ReadUvarint()
	if err != nil {
		return
	}

	var bytes []byte
	bytes, err = reader.ReadBytes(int(length))
	if err != nil {
		return
	}
	difficulty = new(big.Int).SetBytes(bytes)

	return
}

func (chain *Blockchain) saveBlockchain(bucket *bolt.Bucket) error {

	marshal, err := json.Marshal(chain)
	if err != nil {
		return errors.New("Error marshaling chain")
	}

	return bucket.Put([]byte("blockchainInfo"), marshal)
}

func (chain *Blockchain) loadBlockchain() (success bool, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Chain"))
		if reader == nil {
			return nil
		}

		chainData := reader.Get([]byte("blockchainInfo"))
		if chainData == nil {
			return nil
		}

		err = json.Unmarshal(chainData, &chain)
		if err != nil {
			return err
		}
		success = true

		return nil
	})

	return

}
