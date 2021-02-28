package blockchain

import (
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/store"
)

func (chain *Blockchain) init() (err error) {

	chain.Height = 0
	chain.Hash = genesis.GenesisData.Hash
	chain.KernelHash = genesis.GenesisData.KernelHash
	chain.Target = new(big.Int).SetBytes(genesis.GenesisData.Target[:])
	chain.BigTotalDifficulty = new(big.Int).SetUint64(0)

	var tok = token.Token{
		Version:          0,
		Name:             config.NATIVE_TOKEN_NAME,
		Ticker:           config.NATIVE_TOKEN_TICKER,
		Description:      config.NATIVE_TOKEN_DESCRIPTION,
		DecimalSeparator: config.DECIMAL_SEPARATOR,
		CanBurn:          true,
		CanMint:          true,
		Supply:           0,
		MaxSupply:        config.ConvertToUnits(config.MAX_SUPPLY_COINS),
		Key:              config.BURN_PUBLIC_KEY_HASH,
		SupplyKey:        config.BURN_PUBLIC_KEY_HASH,
	}

	if err = store.StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {

		var toks *tokens.Tokens
		if toks, err = tokens.NewTokens(tx); err != nil {
			return
		}
		toks.UpdateToken(config.NATIVE_TOKEN_FULL, &tok)

		toks.Commit()

		return

	}); err != nil {
		return
	}

	return
}

func (chain *Blockchain) computeNextTargetBig(bucket *bolt.Bucket) (*big.Int, error) {

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

	return difficulty.NextTargetBig(deltaTotalDifficulty, deltaTime)
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
