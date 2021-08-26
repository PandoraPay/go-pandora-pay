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
	"pandora-pay/network/websocks/connection/advanced-connection-types"
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

	if err := store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(writer)
		accs := accounts.NewAccounts(writer)

		supply := uint64(0)
		for _, airdrop := range genesis.GenesisData.AirDrops {

			if err = helpers.SafeUint64Add(&supply, airdrop.Amount); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccountEvenEmpty(airdrop.PublicKey, 0); err != nil {
				return
			}

			if airdrop.DelegatedStakePublicKey != nil {
				if err = acc.CreateDelegatedStake(airdrop.Amount, airdrop.DelegatedStakePublicKey, airdrop.DelegatedStakeFee); err != nil {
					return
				}
			} else {
				if err = acc.AddBalance(true, airdrop.Amount, config.NATIVE_TOKEN); err != nil {
					return
				}
			}

			if err = accs.UpdateAccount(airdrop.PublicKey, acc); err != nil {
				return
			}
		}

		tok := &token.Token{
			Version:                  0,
			Name:                     config.NATIVE_TOKEN_NAME,
			Ticker:                   config.NATIVE_TOKEN_TICKER,
			Description:              config.NATIVE_TOKEN_DESCRIPTION,
			DecimalSeparator:         byte(config.DECIMAL_SEPARATOR),
			CanChangePublicKey:       false,
			CanChangeSupplyPublicKey: false,
			CanBurn:                  true,
			CanMint:                  true,
			Supply:                   supply,
			MaxSupply:                config.MAX_SUPPLY_COINS_UNITS,
			UpdatePublicKey:          config.BURN_PUBLIC_KEY,
			SupplyPublicKey:          config.BURN_PUBLIC_KEY,
		}

		if err = toks.CreateToken(config.NATIVE_TOKEN, tok); err != nil {
			return
		}

		chainData.TokensCount = 1

		toks.CommitChanges()
		accs.CommitChanges()

		if err = toks.WriteToStore(); err != nil {
			return
		}
		if err = accs.WriteToStore(); err != nil {
			return
		}

		if err = saveTokensInfo(toks); err != nil {
			return
		}
		if err = toks.Tx.Put("tokenInfo_ByIndex:"+strconv.FormatUint(0, 10), config.NATIVE_TOKEN_FULL); err != nil {
			return
		}

		return

	}); err != nil {
		return nil, err
	}

	chain.ChainData.Store(chainData)
	return chainData, nil
}

func (chain *Blockchain) createNextBlockForForging(chainData *BlockchainData, newWork bool) {

	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL || globals.Arguments["--staking"] == false {
		return
	}

	if chainData == nil {
		chain.mempool.ContinueWork()
	} else {
		chain.mempool.UpdateWork(chainData.Hash, chainData.Height)
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
				MerkleHash:     cryptography.SHA3([]byte{}),
				PrevHash:       chainData.Hash,
				PrevKernelHash: chainData.KernelHash,
				Timestamp:      chainData.Timestamp,
			}
		}

		blk.Forger = make([]byte, cryptography.PublicKeySize)
		blk.DelegatedPublicKey = make([]byte, cryptography.PublicKeySize)
		blk.Signature = make([]byte, cryptography.SignatureSize)

		blk.BloomSerializedNow(blk.SerializeManualToBytes())

		blkComplete := &block_complete.BlockComplete{
			Block: blk,
			Txs:   []*transaction.Transaction{},
		}

		blkComplete.BloomCompleteBySerialized(blkComplete.SerializeManualToBytes())

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

			go func() {
				err := chain.AddBlocks([]*block_complete.BlockComplete{blkComplete}, true, advanced_connection_types.UUID_ALL)
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
