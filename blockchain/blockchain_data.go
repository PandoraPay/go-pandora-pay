package blockchain

import (
	"encoding/hex"
	"errors"
	"math/big"
	difficulty "pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type BlockchainData struct {
	Hash                  helpers.HexBytes `json:"hash"`           //32
	PrevHash              helpers.HexBytes `json:"prevHash"`       //32
	KernelHash            helpers.HexBytes `json:"kernelHash"`     //32
	PrevKernelHash        helpers.HexBytes `json:"prevKernelHash"` //32
	Height                uint64           `json:"height"`
	Timestamp             uint64           `json:"timestamp"`
	Target                *big.Int         `json:"target"`
	BigTotalDifficulty    *big.Int         `json:"bigTotalDifficulty"`
	TransactionsCount     uint64           `json:"transactionsCount"` //count of the number of txs
	TokensCount           uint64           `json:"tokensCount"`       //count of the number of tokens
	AccountsCount         uint64           `json:"accountsCount"`
	ConsecutiveSelfForged uint64           `json:"consecutiveSelfForged"`
}

func (chainData *BlockchainData) computeNextTargetBig(reader store_db_interface.StoreDBTransactionInterface) (*big.Int, error) {

	if config.DIFFICULTY_BLOCK_WINDOW > chainData.Height {
		return chainData.Target, nil
	}

	first := chainData.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp, err := chainData.LoadTotalDifficultyExtra(reader, first+1)
	if err != nil {
		return nil, err
	}

	lastDifficulty := chainData.BigTotalDifficulty
	lastTimestamp := chainData.Timestamp

	deltaTotalDifficulty := new(big.Int).Sub(lastDifficulty, firstDifficulty)
	deltaTime := lastTimestamp - firstTimestamp

	//gui.Log("lastDifficulty", lastDifficulty.String(), "chainData.Height", chainData.Height, "chainData.Timestamp", chainData.Timestamp, "chainData.BigTotalDifficulty", chainData.BigTotalDifficulty.String())
	if deltaTotalDifficulty.Cmp(config.BIG_INT_ZERO) == 0 {
		return nil, errors.New("Delta Difficulty is zero")
	}

	return difficulty.NextTargetBig(deltaTotalDifficulty, deltaTime)
}

func (chainData *BlockchainData) updateChainInfo() {
	gui.GUI.Info2Update("Blocks", strconv.FormatUint(chainData.Height, 10))
	gui.GUI.Info2Update("Chain  Hash", hex.EncodeToString(chainData.Hash))
	gui.GUI.Info2Update("Chain KHash", hex.EncodeToString(chainData.KernelHash))
	gui.GUI.Info2Update("Chain  Diff", chainData.Target.String())
	gui.GUI.Info2Update("TXs", strconv.FormatUint(chainData.TransactionsCount, 10))
}
