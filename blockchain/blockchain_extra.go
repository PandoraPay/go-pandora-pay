package blockchain

import (
	"errors"
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
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
		Hash:               helpers.CloneBytes(genesis.GenesisData.Hash),
		PrevHash:           helpers.CloneBytes(genesis.GenesisData.Hash),
		KernelHash:         helpers.CloneBytes(genesis.GenesisData.KernelHash),
		PrevKernelHash:     helpers.CloneBytes(genesis.GenesisData.KernelHash),
		Target:             new(big.Int).SetBytes(helpers.CloneBytes(genesis.GenesisData.Target)),
		BigTotalDifficulty: new(big.Int).SetUint64(0),
		TransactionsCount:  0,
		AssetsCount:        1,
	}
}

func (chain *Blockchain) initializeNewChain(chainData *BlockchainData, dataStorage *data_storage.DataStorage) (err error) {

	var accs *accounts.Accounts

	if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET); err != nil {
		return
	}

	supply := uint64(0)

	for _, airdrop := range genesis.GenesisData.AirDrops {

		if err = helpers.SafeUint64Add(&supply, airdrop.Amount); err != nil {
			return
		}

		var addr *addresses.Address
		addr, err = addresses.DecodeAddr(airdrop.Address)
		if err != nil {
			return
		}
		if addr.IsIntegratedAmount() || addr.IsIntegratedPaymentID() {
			return errors.New("Amount or PaymentID are integrated there should not be")
		}

		if airdrop.DelegatedStakePublicKey != nil {
			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = dataStorage.PlainAccs.CreatePlainAccount(addr.PublicKey); err != nil {
				return
			}
			if err = plainAcc.CreateDelegatedStake(airdrop.Amount, airdrop.DelegatedStakePublicKey, airdrop.DelegatedStakeFee); err != nil {
				return
			}
			if err = dataStorage.PlainAccs.Update(string(addr.PublicKey), plainAcc); err != nil {
				return
			}
		} else {
			if _, err = dataStorage.Regs.CreateRegistration(addr.PublicKey, addr.Registration); err != nil {
				return
			}
			var acc *account.Account
			if acc, err = accs.CreateAccount(addr.PublicKey); err != nil {
				return
			}
			if err = acc.Balance.AddBalanceUint(airdrop.Amount); err != nil {
				return
			}
			if err = accs.Update(string(addr.PublicKey), acc); err != nil {
				return
			}
		}

	}

	ast := &asset.Asset{
		Version:                  0,
		Name:                     config_coins.NATIVE_ASSET_NAME,
		Ticker:                   config_coins.NATIVE_ASSET_TICKER,
		Description:              config_coins.NATIVE_ASSET_DESCRIPTION,
		DecimalSeparator:         byte(config_coins.DECIMAL_SEPARATOR),
		CanChangePublicKey:       false,
		CanChangeSupplyPublicKey: false,
		CanBurn:                  true,
		CanMint:                  true,
		Supply:                   supply,
		MaxSupply:                config_coins.MAX_SUPPLY_COINS_UNITS,
		UpdatePublicKey:          config_coins.BURN_PUBLIC_KEY,
		SupplyPublicKey:          config_coins.BURN_PUBLIC_KEY,
	}

	if err = dataStorage.Asts.CreateAsset(config_coins.NATIVE_ASSET, ast); err != nil {
		return
	}

	if err = dataStorage.CommitChanges(); err != nil {
		return
	}

	return
}

func (chain *Blockchain) init() (*BlockchainData, error) {

	chainData := chain.createGenesisBlockchainData()

	if err := store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.CreateDataStorage(writer)

		if err = chain.initializeNewChain(chainData, dataStorage); err != nil {
			return
		}

		if config.SEED_WALLET_NODES_INFO {
			if err = saveAssetsInfo(dataStorage.Asts); err != nil {
				return
			}
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
		blk.DelegatedStakePublicKey = make([]byte, cryptography.PublicKeySize)
		blk.Signature = make([]byte, cryptography.SignatureSize)

		blk.BloomSerializedNow(blk.SerializeManualToBytes())

		blkComplete := &block_complete.BlockComplete{
			Block: blk,
			Txs:   []*transaction.Transaction{},
		}

		if err = blkComplete.BloomCompleteBySerialized(blkComplete.SerializeManualToBytes()); err != nil {
			return
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

			recovery.SafeGo(func() {
				err := chain.AddBlocks([]*block_complete.BlockComplete{blkComplete}, true, advanced_connection_types.UUID_ALL)
				if err == nil {
					gui.GUI.Info("Block was forged! " + strconv.FormatUint(blkComplete.Block.Height, 10))
				} else {
					gui.GUI.Error("Error forging block "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
				}
			})

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
			}

		}
	})

}
