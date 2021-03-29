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

func (chain *Blockchain) createGenesisBlockchainData() *BlockchainData {
	return &BlockchainData{
		Height:             0,
		Hash:               genesis.GenesisData.Hash,
		PrevHash:           genesis.GenesisData.Hash,
		KernelHash:         genesis.GenesisData.KernelHash,
		PrevKernelHash:     genesis.GenesisData.KernelHash,
		Target:             new(big.Int).SetBytes(genesis.GenesisData.Target),
		BigTotalDifficulty: new(big.Int).SetUint64(0),
	}
}

func (chain *Blockchain) init() (err error) {

	chainData := chain.createGenesisBlockchainData()
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

	return store.StoreBlockchain.DB.Update(func(boltTx *bolt.Tx) (err error) {

		toks := tokens.NewTokens(boltTx)
		if err = toks.CreateToken(config.NATIVE_TOKEN, &tok); err != nil {
		}

		toks.Commit()
		if err = toks.WriteToStore(); err != nil {
			return
		}

		return

	})
}

func (chain *Blockchain) createNextBlockForForging() {

	chain.RLock()
	chainData := chain.GetChainData()
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
	blk.Forger = make([]byte, 20)
	blk.Signature = make([]byte, 65)

	blkComplete := &block_complete.BlockComplete{
		Block: blk,
	}

	chain.RUnlock()

	chain.forging.ForgingNewWork(blkComplete, target)
}

func (chain *Blockchain) initForging() {

	go func() {

		var err error
		for {

			blkComplete, ok := <-chain.forging.SolutionCn
			if !ok {
				return
			}

			blkComplete.Block.Bloom = nil
			blkComplete.Bloom = nil

			if err = blkComplete.BloomAll(); err != nil {
				gui.Error("Error blooming forged blkComplete", err)
				chain.mempool.RestartWork()
				continue
			}

			array := []*block_complete.BlockComplete{blkComplete}

			err := chain.AddBlocks(array, true)
			if err == nil {
				gui.Info("Block was forged! " + strconv.FormatUint(blkComplete.Block.Height, 10))
			} else if err != nil {
				gui.Error("Error forging block "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
				chain.mempool.RestartWork()
			}

		}

	}()

	go chain.createNextBlockForForging()

}
