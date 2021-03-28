package blockchain

import (
	"encoding/hex"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

type BlockchainData struct {
	Hash                  []byte //32
	PrevHash              []byte //32
	KernelHash            []byte //32
	PrevKernelHash        []byte //32
	Height                uint64
	Timestamp             uint64
	Target                *big.Int
	BigTotalDifficulty    *big.Int
	Transactions          uint64 //count of the number of txs
	ConsecutiveSelfForged uint64
}

func (chainData *BlockchainData) loadTotalDifficultyExtra(bucket *bolt.Bucket, height uint64) (difficulty *big.Int, timestamp uint64, err error) {
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

func (chainData *BlockchainData) computeNextTargetBig(bucket *bolt.Bucket) (*big.Int, error) {

	if config.DIFFICULTY_BLOCK_WINDOW > chainData.Height {
		return chainData.Target, nil
	}

	first := chainData.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp, err := chainData.loadTotalDifficultyExtra(bucket, first)
	if err != nil {
		return nil, err
	}

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
