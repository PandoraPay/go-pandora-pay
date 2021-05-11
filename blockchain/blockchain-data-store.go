package blockchain

import (
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/helpers"
	"strconv"
)

//chain must be locked before
func (chainData *BlockchainData) saveTotalDifficultyExtra(bucket *bolt.Bucket) error {
	key := []byte("totalDifficulty" + strconv.FormatUint(chainData.Height, 10))

	writer := helpers.NewBufferWriter()
	writer.WriteUvarint(chainData.Timestamp)

	bytes := chainData.BigTotalDifficulty.Bytes()
	writer.WriteUvarint(uint64(len(bytes)))
	writer.Write(bytes)

	return bucket.Put(key, writer.Bytes())
}

func (chainData *BlockchainData) LoadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64, err error) {
	if height < 0 {
		err = errors.New("height is invalid")
		return
	}
	key := []byte("totalDifficulty" + strconv.FormatUint(height, 10))

	buf := bucket.Get(key)
	if buf == nil {
		err = errors.New("Couldn't ready difficulty from DB")
		return
	}

	reader := helpers.NewBufferReader(buf)
	if timestamp, err = reader.ReadUvarint(); err != nil {
		return
	}

	length, err := reader.ReadUvarint()
	if err != nil {
		return
	}

	bytes, err := reader.ReadBytes(int(length))
	if err != nil {
		return
	}

	difficulty = new(big.Int).SetBytes(bytes)

	return
}

func (chainData *BlockchainData) loadBlockchainInfo(bucket *bolt.Bucket, height uint64) error {
	chainInfoData := bucket.Get([]byte("blockchainInfo_" + strconv.FormatUint(height, 10)))
	if chainInfoData == nil {
		return errors.New("Chain not found")
	}
	return json.Unmarshal(chainInfoData, chainData)
}

func (chainData *BlockchainData) saveBlockchainInfo(bucket *bolt.Bucket) (err error) {
	var data []byte
	if data, err = json.Marshal(chainData); err != nil {
		return
	}

	return bucket.Put([]byte("blockchainInfo_"+strconv.FormatUint(chainData.Height, 10)), data)
}

func (chainData *BlockchainData) saveBlockchain(bucket *bolt.Bucket) error {
	marshal, err := json.Marshal(chainData)
	if err != nil {
		return err
	}

	return bucket.Put([]byte("blockchainInfo"), marshal)
}
