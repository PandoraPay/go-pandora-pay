package blockchain

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (chain *Blockchain) LoadBlockFromHashSilent(hash helpers.Hash) (blk *block.Block, err error) {

	err = store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		defer func() {
			if err2 := recover(); err2 != nil {
				err = helpers.ConvertRecoverError(err2)
			}
		}()

		reader := tx.Bucket([]byte("Chain"))
		if reader == nil {
			return nil
		}

		blk = chain.loadBlock(reader, hash)

		return err
	})

	return
}

func (chain *Blockchain) loadBlock(bucket *bolt.Bucket, hash helpers.Hash) (blk *block.Block) {

	key := []byte("blockHash")
	key = append(key, hash[:]...)

	blockData := bucket.Get(key)
	if blockData == nil {
		return
	}

	blk = &block.Block{}

	reader := helpers.NewBufferReader(blockData)
	blk.Deserialize(reader)

	return
}

func (chain *Blockchain) saveBlock(bucket *bolt.Bucket, blkComplete *block.BlockComplete, hash helpers.Hash) {

	key := append([]byte("blockHash"), hash[:]...)

	if err := bucket.Put(key, blkComplete.Block.Serialize()); err != nil {
		panic(err)
	}

	key = []byte("blockHeight" + strconv.FormatUint(blkComplete.Block.Height, 10))
	if err := bucket.Put(key, hash[:]); err != nil {
		panic(err)
	}

}

func (chain *Blockchain) loadBlockHash(bucket *bolt.Bucket, height uint64) helpers.Hash {

	if height < 0 || height > chain.Height {
		panic("Height is invalid")
	}

	key := []byte("blockHeight" + strconv.FormatUint(height, 10))
	return *helpers.ConvertHash(bucket.Get(key))
}

func (chain *Blockchain) saveTotalDifficultyExtra(bucket *bolt.Bucket) {
	key := []byte("totalDifficulty" + strconv.FormatUint(chain.Height, 10))

	writer := helpers.NewBufferWriter()
	writer.WriteUvarint(chain.Timestamp)

	bytes := chain.BigTotalDifficulty.Bytes()
	writer.WriteUvarint(uint64(len(bytes)))
	writer.Write(bytes)

	if err := bucket.Put(key, writer.Bytes()); err != nil {
		panic(err)
	}
}

func (chain *Blockchain) loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64) {
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

	if err := bucket.Put([]byte("blockchainInfo"), marshal); err != nil {
		panic(err)
	}
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

		if err = json.Unmarshal(chainData, &chain); err != nil {
			return err
		}
		success = true

		return nil
	})

	return

}
