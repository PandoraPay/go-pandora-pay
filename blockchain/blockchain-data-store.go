package blockchain

import (
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/helpers"
	"strconv"
)

//chain must be locked before
func (chainData *BlockchainData) saveTotalDifficultyExtra(bucket *bolt.Bucket) {
	key := []byte("totalDifficulty" + strconv.FormatUint(chainData.Height, 10))

	writer := helpers.NewBufferWriter()
	writer.WriteUvarint(chainData.Timestamp)

	bytes := chainData.BigTotalDifficulty.Bytes()
	writer.WriteUvarint(uint64(len(bytes)))
	writer.Write(bytes)

	bucket.Put(key, writer.Bytes())
}

func (chainData *BlockchainData) saveBlockchain(bucket *bolt.Bucket) {
	marshal, err := json.Marshal(chainData)
	if err != nil {
		panic(err)
	}

	bucket.Put([]byte("blockchainInfo"), marshal)
}
