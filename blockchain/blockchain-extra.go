package blockchain

import (
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
)

func (chain *Blockchain) computeNextDifficultyBig(bucket *bolt.Bucket) (*big.Int, error) {

	if config.DIFFICULTY_BLOCK_WINDOW > chain.Height {
		return chain.Target, nil
	}

	first := chain.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp, err := loadTotalDifficultyExtra(bucket, first)
	if err != nil {
		return nil, err
	}

	lastDifficulty := chain.BigTotalDifficulty
	lastTimestamp := chain.Timestamp

	deltaTotalDifficulty := new(big.Int).Sub(lastDifficulty, firstDifficulty)
	deltaTime := lastTimestamp - firstTimestamp

	return difficulty.NextDifficultyBig(deltaTotalDifficulty, deltaTime)
}

func (chain *Blockchain) createNextBlockComplete() (blkComplete *block.BlockComplete, err error) {

	var blk *block.Block
	if chain.Height == 0 {
		if blk, err = genesis.CreateNewGenesisBlock(); err != nil {
			return
		}
	} else {

		chain.RLock()

		var blockHeader = block.BlockHeader{
			Version: 0,
			Height:  chain.Height,
		}

		blk = &block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     crypto.SHA3Hash([]byte{}),
			PrevHash:       chain.Hash,
			PrevKernelHash: chain.KernelHash,
			Timestamp:      chain.Timestamp,
		}

		chain.RUnlock()

	}

	blkComplete = &block.BlockComplete{
		Block: blk,
	}

	return
}

func (chain *Blockchain) createBlockForForging() {

	var err error

	var nextBlock *block.BlockComplete
	if nextBlock, err = Chain.createNextBlockComplete(); err != nil {
		gui.Error("Error creating next block", err)
	}

	forging.Forging.RestartForgingWorkers(nextBlock, chain.Target)
}
