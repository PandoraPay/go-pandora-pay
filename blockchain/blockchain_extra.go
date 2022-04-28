package blockchain

import (
	"errors"
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/blockchain_types"
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
	"pandora-pay/config/config_forging"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

func (chain *Blockchain) GetChainData() *BlockchainData {
	return chain.ChainData.Load()
}

func (chain *Blockchain) GetChainDataUpdate() *BlockchainDataUpdate {
	return &BlockchainDataUpdate{chain.ChainData.Load(), chain.Sync.GetSyncData()}
}

func (chain *Blockchain) createGenesisBlockchainData() *BlockchainData {
	return &BlockchainData{
		helpers.CloneBytes(genesis.GenesisData.Hash),
		helpers.CloneBytes(genesis.GenesisData.Hash),
		helpers.CloneBytes(genesis.GenesisData.KernelHash),
		helpers.CloneBytes(genesis.GenesisData.KernelHash),
		0,
		0,
		new(big.Int).SetBytes(helpers.CloneBytes(genesis.GenesisData.Target)),
		new(big.Int).SetUint64(0),
		0,
		0,
		0,
		0,
		0,
	}
}

func (chain *Blockchain) initializeNewChain(chainData *BlockchainData, dataStorage *data_storage.DataStorage) (err error) {

	gui.GUI.Info("Initializing New Chain")

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
		if addr.IsIntegratedAmount() || addr.IsIntegratedPaymentID() || addr.IsIntegratedPaymentAsset() {
			return errors.New("Amount, PaymentID or IntegratedPaymentAsset are not allowed in the airdrop address")
		}

		var accs *accounts.Accounts
		var acc *account.Account
		var plainAcc *plain_account.PlainAccount

		if len(airdrop.DelegatedStakePublicKey) == cryptography.PublicKeySize {
			if plainAcc, err = dataStorage.CreatePlainAccount(addr.PublicKeyHash); err != nil {
				return
			}

			amount := airdrop.Amount

			if amount > config_coins.ConvertToUnitsUint64Forced(500) {
				if err = helpers.SafeUint64Sub(&amount, config_coins.ConvertToUnitsUint64Forced(500)); err != nil {
					return
				}

				if accs, acc, err = dataStorage.CreateAccount(config_coins.NATIVE_ASSET_FULL, addr.PublicKeyHash); err != nil {
					return
				}
				acc.Balance = config_coins.ConvertToUnitsUint64Forced(500)
				if err = accs.Update(string(addr.PublicKeyHash), acc); err != nil {
					return
				}
			}
			if err = plainAcc.AddStakeAvailable(true, amount); err != nil {
				return
			}

			if err = plainAcc.DelegatedStake.CreateDelegatedStake(0, airdrop.DelegatedStakePublicKey, airdrop.DelegatedStakeFee); err != nil {
				return
			}
			if err = dataStorage.PlainAccs.Update(string(addr.PublicKeyHash), plainAcc); err != nil {
				return
			}

		} else {
			if accs, acc, err = dataStorage.CreateAccount(config_coins.NATIVE_ASSET_FULL, addr.PublicKeyHash); err != nil {
				return
			}
			acc.Balance = airdrop.Amount
			if err = accs.Update(string(addr.PublicKeyHash), acc); err != nil {
				return
			}
		}

	}

	ast := &asset.Asset{
		nil,
		nil,
		0,
		0,
		false,
		false,
		false,
		false,
		false,
		false,
		false,
		byte(config_coins.DECIMAL_SEPARATOR),
		config_coins.MAX_SUPPLY_COINS_UNITS,
		supply,
		config_coins.BURN_PUBLIC_KEY,
		config_coins.BURN_PUBLIC_KEY,
		config_coins.NATIVE_ASSET_NAME,
		config_coins.NATIVE_ASSET_TICKER,
		config_coins.NATIVE_ASSET_IDENTIFICATION,
		config_coins.NATIVE_ASSET_DESCRIPTION,
		nil,
	}

	if err = dataStorage.Asts.CreateAsset(config_coins.NATIVE_ASSET_FULL, ast); err != nil {
		return
	}

	if err = dataStorage.CommitChanges(); err != nil {
		return
	}

	chainData.AssetsCount = dataStorage.Asts.Count
	chainData.AccountsCount = dataStorage.PlainAccs.Count

	return
}

func (chain *Blockchain) init() (*BlockchainData, error) {

	chainData := chain.createGenesisBlockchainData()

	if err := store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(writer)

		if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
			if err = chain.initializeNewChain(chainData, dataStorage); err != nil {
				return
			}
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

	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL {
		return
	}

	if chainData == nil {
		chain.mempool.ContinueWork()
	} else {
		chain.mempool.UpdateWork(chainData.Hash, chainData.Height)
	}

	if !config_forging.FORGING_ENABLED {
		return
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

		blk.Forger = make([]byte, cryptography.PublicKeyHashSize)
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
			config_stake.GetRequiredStake(blkComplete.Height),
		}

	}

}

func (chain *Blockchain) InitForging() {

	recovery.SafeGo(func() {

		for {

			solution, ok := <-chain.ForgingSolutionCn
			if !ok {
				return
			}

			kernelHash, err := chain.AddBlocks([]*block_complete.BlockComplete{solution.BlkComplete}, true, advanced_connection_types.UUID_ALL)

			solution.Done <- &blockchain_types.BlockchainSolutionAnswer{
				err,
				kernelHash,
			}
		}

	})

	recovery.SafeGo(func() {

		updateNewSyncCn := chain.Sync.UpdateSyncMulticast.AddListener()
		defer chain.Sync.UpdateSyncMulticast.RemoveChannel(updateNewSyncCn)

		for {

			newSync := <-updateNewSyncCn

			if newSync.Started {
				chain.createNextBlockForForging(chain.GetChainData(), true)
				break
			}

		}
	})

}
