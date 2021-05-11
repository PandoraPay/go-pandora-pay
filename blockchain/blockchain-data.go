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
	Hash                  helpers.HexBytes //32
	PrevHash              helpers.HexBytes //32
	KernelHash            helpers.HexBytes //32
	PrevKernelHash        helpers.HexBytes //32
	Height                uint64
	Timestamp             uint64
	Target                *big.Int
	BigTotalDifficulty    *big.Int
	Transactions          uint64 //count of the number of txs
	ConsecutiveSelfForged uint64
}

func (chainData *BlockchainData) computeNextTargetBig(bucket *bolt.Bucket) (*big.Int, error) {

	if config.DIFFICULTY_BLOCK_WINDOW > chainData.Height {
		return chainData.Target, nil
	}

	first := chainData.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp, err := chainData.LoadTotalDifficultyExtra(bucket, first)
	if err != nil {
		return nil, err
	}

	lastDifficulty := chainData.BigTotalDifficulty
	lastTimestamp := chainData.Timestamp

	deltaTotalDifficulty := new(big.Int).Sub(lastDifficulty, firstDifficulty)
	deltaTime := lastTimestamp - firstTimestamp

	gui.Log("lastDifficulty", lastDifficulty.String(), "chainData.Height", chainData.Height, "chainData.Timestamp", chainData.Timestamp, "chainData.BigTotalDifficulty", chainData.BigTotalDifficulty.String())
	if deltaTotalDifficulty.Cmp(config.BIG_INT_ZERO) == 0 {
		gui.Error("ERRROR!!!", lastDifficulty.String())
		return nil, errors.New("Delta Difficulty is zero")
	}

	return difficulty.NextTargetBig(deltaTotalDifficulty, deltaTime)
}

func (chainData *BlockchainData) updateChainInfo() {
	gui.Info2Update("Blocks", strconv.FormatUint(chainData.Height, 10))
	gui.Info2Update("Chain  Hash", hex.EncodeToString(chainData.Hash))
	gui.Info2Update("Chain KHash", hex.EncodeToString(chainData.KernelHash))
	gui.Info2Update("Chain  Diff", chainData.Target.String())
	gui.Info2Update("TXs", strconv.FormatUint(chainData.Transactions, 10))
}
