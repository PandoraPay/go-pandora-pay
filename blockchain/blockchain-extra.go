package blockchain

import (
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

func (chain *Blockchain) init() {

	chain.Height = 0
	chain.Hash = genesis.GenesisData.Hash
	chain.PrevHash = genesis.GenesisData.Hash
	chain.KernelHash = genesis.GenesisData.KernelHash
	chain.PrevKernelHash = genesis.GenesisData.KernelHash
	chain.Target = new(big.Int).SetBytes(genesis.GenesisData.Target)
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

	if err := store.StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {

		toks := tokens.NewTokens(tx)
		toks.CreateToken(config.NATIVE_TOKEN, &tok)

		toks.Commit()
		toks.WriteToStore()

		return

	}); err != nil {
		panic(err)
	}
}

func (chain *Blockchain) computeNextTargetBig(bucket *bolt.Bucket) *big.Int {

	if config.DIFFICULTY_BLOCK_WINDOW > chain.Height {
		return chain.Target
	}

	first := chain.Height - config.DIFFICULTY_BLOCK_WINDOW

	firstDifficulty, firstTimestamp := chain.loadTotalDifficultyExtra(bucket, first)

	lastDifficulty := chain.BigTotalDifficulty
	lastTimestamp := chain.Timestamp

	deltaTotalDifficulty := new(big.Int).Sub(lastDifficulty, firstDifficulty)
	deltaTime := lastTimestamp - firstTimestamp

	return difficulty.NextTargetBig(deltaTotalDifficulty, deltaTime)
}

//make sure that chain is RLocked
func (chain *Blockchain) createNextBlockComplete() (blkComplete *block.BlockComplete, err error) {

	chain.RLock()
	defer chain.RUnlock()

	return
}

func (chain *Blockchain) createNextBlockForForging() {

	chain.RLock()
	target := chain.Target

	var blk *block.Block
	var err error
	if chain.Height == 0 {
		if blk, err = genesis.CreateNewGenesisBlock(); err != nil {
			chain.RUnlock()
			gui.Error("Error creating next block", err)
			return
		}
	} else {

		var blockHeader = block.BlockHeader{
			Version: 0,
			Height:  chain.Height,
		}

		blk = &block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     cryptography.SHA3Hash([]byte{}),
			PrevHash:       chain.Hash,
			PrevKernelHash: chain.KernelHash,
			Timestamp:      chain.Timestamp,
		}

	}
	blk.DelegatedPublicKey = make([]byte, 33)
	blk.Forger = make([]byte, 20)
	blk.Signature = make([]byte, 65)

	blkComplete := &block.BlockComplete{
		Block: blk,
	}

	chain.RUnlock()

	chain.forging.RestartForgingWorkers(blkComplete, target)
}

func (chain *Blockchain) updateChainInfo() {
	gui.Info2Update("Blocks", strconv.FormatUint(chain.Height, 10))
	gui.Info2Update("Chain  Hash", hex.EncodeToString(chain.Hash))
	gui.Info2Update("Chain KHash", hex.EncodeToString(chain.KernelHash))
	gui.Info2Update("Chain  Diff", chain.Target.String())
	gui.Info2Update("TXs", strconv.FormatUint(chain.Transactions, 10))
}
