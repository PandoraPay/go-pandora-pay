package blockchain

import (
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

type BlockchainData struct {
	Hash               []byte //32
	PrevHash           []byte //32
	KernelHash         []byte //32
	PrevKernelHash     []byte //32
	Height             uint64
	Timestamp          uint64
	Target             *big.Int
	BigTotalDifficulty *big.Int
	Transactions       uint64 //count of the number of txs
}

func (chainData *BlockchainData) loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64) {
	if height < 0 {
		panic("height is invalid")
	}
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

func (chainData *BlockchainData) computeNextTargetBig(bucket *bolt.Bucket) *big.Int {

	if config.DIFFICULTY_BLOCK_WINDOW > chainData.Height {
		return chainData.Target
	}

	first := chainData.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp := chainData.loadTotalDifficultyExtra(bucket, first)

	lastDifficulty := chainData.BigTotalDifficulty
	lastTimestamp := chainData.Timestamp

	deltaTotalDifficulty := new(big.Int).Sub(lastDifficulty, firstDifficulty)
	deltaTime := lastTimestamp - firstTimestamp

	return difficulty.NextTargetBig(deltaTotalDifficulty, deltaTime)
}

func (chainData *BlockchainData) updateChainInfo() {
	gui.Info2Update("Blocks", strconv.FormatUint(chainData.Height, 10))
	gui.Info2Update("Chain  Hash", hex.EncodeToString(chainData.Hash))
	gui.Info2Update("Chain KHash", hex.EncodeToString(chainData.KernelHash))
	gui.Info2Update("Chain  Diff", chainData.Target.String())
	gui.Info2Update("TXs", strconv.FormatUint(chainData.Transactions, 10))
}
