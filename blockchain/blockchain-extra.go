package blockchain

import (
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
	"sync/atomic"
	"unsafe"
)

func (chain *Blockchain) GetChainData() *BlockchainData {
	pointer := atomic.LoadPointer(&chain.ChainData)
	return (*BlockchainData)(pointer)
}

func (chain *Blockchain) init() {

	chainData := &BlockchainData{
		Height:             0,
		Hash:               genesis.GenesisData.Hash,
		PrevHash:           genesis.GenesisData.Hash,
		KernelHash:         genesis.GenesisData.KernelHash,
		PrevKernelHash:     genesis.GenesisData.KernelHash,
		Target:             new(big.Int).SetBytes(genesis.GenesisData.Target),
		BigTotalDifficulty: new(big.Int).SetUint64(0),
	}
	atomic.StorePointer(&chain.ChainData, unsafe.Pointer(chainData))

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

	if err := store.StoreBlockchain.DB.Update(func(boltTx *bolt.Tx) (err error) {

		toks := tokens.NewTokens(boltTx)
		toks.CreateToken(config.NATIVE_TOKEN, &tok)

		toks.Commit()
		toks.WriteToStore()

		return

	}); err != nil {
		panic(err)
	}
}

func (chain *Blockchain) createNextBlockForForging() {

	chain.RLock()
	chainData := (*BlockchainData)(atomic.LoadPointer(&chain.ChainData))
	target := chainData.Target

	var blk *block.Block
	var err error
	if chainData.Height == 0 {
		if blk, err = genesis.CreateNewGenesisBlock(); err != nil {
			chain.RUnlock()
			gui.Error("Error creating next block", err)
			return
		}
	} else {

		var blockHeader = block.BlockHeader{
			Version: 0,
			Height:  chainData.Height,
		}

		blk = &block.Block{
			BlockHeader:    blockHeader,
			MerkleHash:     cryptography.SHA3Hash([]byte{}),
			PrevHash:       chainData.Hash,
			PrevKernelHash: chainData.KernelHash,
			Timestamp:      chainData.Timestamp,
		}

	}
	blk.DelegatedPublicKeyHash = make([]byte, 20)
	blk.Forger = make([]byte, 20)
	blk.Signature = make([]byte, 65)

	blkComplete := &block_complete.BlockComplete{
		Block: blk,
	}

	chain.RUnlock()

	chain.forging.RestartForgingWorkers(blkComplete, target)
}

func (chain *Blockchain) initForging() {

	go func() {

		for {

			blkComplete := <-chain.forging.SolutionChannel
			blkComplete.BloomNow()
			blkComplete.Block.BloomNow()

			array := []*block_complete.BlockComplete{blkComplete}

			result, err := chain.AddBlocks(array, true)
			if err == nil && result {
				gui.Info("Block was forged! " + strconv.FormatUint(blkComplete.Block.Height, 10))
			} else if err != nil {
				gui.Error("Error forging block "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
				chain.mempool.RestartWork()
			} else {
				gui.Warning("Forging block  return false "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
			}

		}

	}()

	go chain.createNextBlockForForging()

}
