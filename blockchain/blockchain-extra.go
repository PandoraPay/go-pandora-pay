package blockchain

import (
	"math/big"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	blockchain_sync "pandora-pay/blockchain/blockchain-sync"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block-complete"
	forging_block_work "pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/recovery"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func (chain *Blockchain) GetChainData() *BlockchainData {
	return chain.ChainData.Load().(*BlockchainData)
}

func (chain *Blockchain) GetChainDataUpdate() *BlockchainDataUpdate {
	chainData := chain.ChainData.Load().(*BlockchainData)
	return &BlockchainDataUpdate{chainData, chain.Sync.GetSyncData()}
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

func (chain *Blockchain) init() (*BlockchainData, error) {

	chainData := chain.createGenesisBlockchainData()
	chain.ChainData.Store(chainData)

	return chainData, store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(writer)
		accs := accounts.NewAccounts(writer)

		supply := uint64(0)
		for _, airdrop := range genesis.GenesisData.AirDrops {

			if err = helpers.SafeUint64Add(&supply, airdrop.Amount); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccountEvenEmpty(airdrop.PublicKeyHash, 0); err != nil {
				return
			}

			if airdrop.DelegatedStakePublicKeyHash != nil {
				if err = acc.CreateDelegatedStake(airdrop.Amount, airdrop.DelegatedStakePublicKeyHash); err != nil {
					return
				}
			} else {
				if err = acc.AddBalance(true, airdrop.Amount, config.NATIVE_TOKEN); err != nil {
					return
				}
			}

			if err = accs.UpdateAccount(airdrop.PublicKeyHash, acc); err != nil {
				return
			}
		}

		maxSupply, err := config.ConvertToUnitsUint64(config.MAX_SUPPLY_COINS)
		if err != nil {
			panic(err)
		}

		tok := token.Token{
			Version:          0,
			Name:             config.NATIVE_TOKEN_NAME,
			Ticker:           config.NATIVE_TOKEN_TICKER,
			Description:      config.NATIVE_TOKEN_DESCRIPTION,
			DecimalSeparator: byte(config.DECIMAL_SEPARATOR),
			CanBurn:          true,
			CanMint:          true,
			Supply:           supply,
			MaxSupply:        maxSupply,
			Key:              config.BURN_PUBLIC_KEY_HASH,
			SupplyKey:        config.BURN_PUBLIC_KEY_HASH,
		}

		if err = toks.CreateToken(config.NATIVE_TOKEN, &tok); err != nil {
			return
		}

		toks.CommitChanges()
		accs.CommitChanges()

		if err = toks.WriteToStore(); err != nil {
			return
		}
		if err = accs.WriteToStore(); err != nil {
			return
		}

		return

	})

}

func (chain *Blockchain) createNextBlockForForging(chainData *BlockchainData, newWork bool) {

	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL || globals.Arguments["--staking"] == false {
		return
	}

	if chainData != nil {
		chain.mempool.UpdateWork(chainData.Hash, chainData.Height)
	} else {
		chain.mempool.ContinueWork()
	}

	if newWork {

		if chainData == nil {
			chainData = chain.GetChainData()
		}

		target := chainData.Target

		var blk *block.Block
		var err error
		if chainData.Height == 0 {
			if blk, err = genesis.CreateNewGenesisBlock(); err != nil {
				gui.GUI.Error("Error creating next block", err)
				return
			}
		} else {
			blk = &block.Block{
				BlockHeader: &block.BlockHeader{
					Version: 0,
					Height:  chainData.Height,
				},
				MerkleHash:     cryptography.SHA3Hash([]byte{}),
				PrevHash:       chainData.Hash,
				PrevKernelHash: chainData.KernelHash,
				Timestamp:      chainData.Timestamp,
			}
		}
		blk.Forger = make([]byte, cryptography.PublicKeyHashHashSize)
		blk.Signature = make([]byte, cryptography.SignatureSize)

		blkComplete := &block_complete.BlockComplete{
			Block: blk,
			Txs:   []*transaction.Transaction{},
		}

		writer := helpers.NewBufferWriter()
		blk.SerializeForForging(writer)

		chain.NextBlockCreatedCn <- &forging_block_work.ForgingWork{
			blkComplete,
			writer.Bytes(),
			blkComplete.Timestamp,
			blkComplete.Height,
			target,
		}

	} else {

		if chainData != nil {
			chain.NextBlockCreatedCn <- nil
		}

	}

}

func (chain *Blockchain) InitForging() {

	recovery.SafeGo(func() {

		var err error
		for {

			blkComplete, ok := <-chain.ForgingSolutionCn
			if !ok {
				return
			}

			blkComplete.Block.Bloom = nil
			blkComplete.Bloom = nil

			if err = blkComplete.BloomAll(); err != nil {
				gui.GUI.Error("Error blooming forged blkComplete", err)
				continue
			}

			array := []*block_complete.BlockComplete{blkComplete}

			go func() {
				err := chain.AddBlocks(array, true)
				if err == nil {
					gui.GUI.Info("Block was forged! " + strconv.FormatUint(blkComplete.Block.Height, 10))
				} else {
					gui.GUI.Error("Error forging block "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
				}
			}()

		}

	})

	recovery.SafeGo(func() {

		updateNewSyncCn := chain.Sync.UpdateSyncMulticast.AddListener()
		defer chain.Sync.UpdateSyncMulticast.RemoveChannel(updateNewSyncCn)

		for {

			newSyncDataReceived, ok := <-updateNewSyncCn
			if !ok {
				break
			}

			newSyncData := newSyncDataReceived.(*blockchain_sync.BlockchainSyncData)
			if newSyncData.Sync {
				chain.createNextBlockForForging(chain.GetChainData(), true)
				break
			}

		}
	})

}
