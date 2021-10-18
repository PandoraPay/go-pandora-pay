package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_coins"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
	"sync"
)

type TransactionsBuilder struct {
	wallet  *wallet.Wallet
	mempool *mempool.Mempool
	chain   *blockchain.Blockchain
	lock    *sync.Mutex //TODO replace sync.Mutex with a snyc.Map in order to optimize the transactions creation
}

func (builder *TransactionsBuilder) getNonce(nonce uint64, publicKey []byte, accNonce uint64) uint64 {
	if nonce != 0 {
		return nonce
	}
	return builder.mempool.GetNonce(publicKey, accNonce)
}

func (builder *TransactionsBuilder) DeriveDelegatedStake(nonce uint64, addressPublicKey []byte) (*wallet_address.WalletAddressDelegatedStake, error) {

	var accNonce uint64
	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		plainAccs := plain_accounts.NewPlainAccounts(reader)

		var plainAcc *plain_account.PlainAccount
		if plainAcc, err = plainAccs.GetPlainAccount(addressPublicKey, chainHeight); err != nil {
			return
		}
		if plainAcc == nil {
			return errors.New("Account doesn't exist")
		}

		return
	}); err != nil {
		return nil, err
	}

	nonce = builder.getNonce(nonce, addressPublicKey, accNonce)

	builder.wallet.RLock()
	defer builder.wallet.RUnlock()

	addr := builder.wallet.GetWalletAddressByPublicKey(addressPublicKey)
	if addr == nil {
		return nil, errors.New("Wallet was not found")
	}

	return addr.DeriveDelegatedStake(uint32(nonce))
}

func (builder *TransactionsBuilder) convertFloatAmounts(amounts []float64, ast *asset.Asset) ([]uint64, error) {

	var err error

	amountsFinal := make([]uint64, len(amounts))
	for i := range amounts {
		if err != nil {
			return nil, err
		}
		if amountsFinal[i], err = ast.ConvertToUnits(amounts[i]); err != nil {
			return nil, err
		}
	}

	return amountsFinal, nil
}

func (builder *TransactionsBuilder) getWalletAddresses(from []string) ([]*wallet_address.WalletAddress, error) {

	fromWalletAddress := make([]*wallet_address.WalletAddress, len(from))
	var err error

	for i, fromAddress := range from {
		if fromWalletAddress[i], err = builder.wallet.GetWalletAddressByEncodedAddress(fromAddress); err != nil {
			return nil, err
		}
		if fromWalletAddress[i].PrivateKey == nil {
			return nil, errors.New("Can't be used for transactions as the private key is missing")
		}
	}

	return fromWalletAddress, nil
}

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(from string, nonce uint64, unstakeAmount float64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	statusCallback("Converting Floats to Numbers")

	unstakeAmountFinal, err := config_coins.ConvertToUnits(unstakeAmount)
	if err != nil {
		return nil, err
	}

	feeFinal := &wizard.TransactionsWizardFee{}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		asts := assets.NewAssets(reader)

		ast, err := asts.GetAsset(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}
		if ast == nil {
			return errors.New("Asset was not found")
		}

		if feeFinal, err = fee.convertToWizardFee(ast); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateUnstakeTx(from, nonce, unstakeAmountFinal, data, feeFinal, propagateTx, awaitAnswer, awaitBroadcast, false, statusCallback)
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, nonce, unstakeAmount uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	statusCallback("Wallet Addresses Found")

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var plainAcc *plain_account.PlainAccount
	var chainHeight uint64

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		plainAccs := plain_accounts.NewPlainAccounts(reader)

		if plainAcc, err = plainAccs.GetPlainAccount(fromWalletAddresses[0].PublicKey, chainHeight); err != nil {
			return
		}
		if plainAcc == nil {
			return errors.New("Account doesn't exist")
		}

		availableStake, err := plainAcc.ComputeDelegatedStakeAvailable(chainHeight)
		if err != nil {
			return
		}

		if availableStake < unstakeAmount {
			return errors.New("You don't have enough staked coins")
		}

		return
	}); err != nil {
		return nil, err
	}

	statusCallback("Balances checked")

	nonce = builder.getNonce(nonce, fromWalletAddresses[0].PublicKey, plainAcc.Nonce)
	statusCallback("Getting Nonce from Mempool")

	if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddresses[0].PrivateKey.Key, unstakeAmount, data, fee, false, statusCallback); err != nil {
		return nil, err
	}
	statusCallback("Transaction Created")

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TransactionsBuilder) CreateUpdateDelegateTx_Float(from string, nonce uint64, delegatedStakingNewPublicKey []byte, delegatedStakingNewFee uint64, delegatedStakingClaimAmount float64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var finalFee *wizard.TransactionsWizardFee
	var finalDelegatedStakingClaimAmount uint64

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		asts := assets.NewAssets(reader)

		ast, err := asts.GetAsset(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}
		if ast == nil {
			return errors.New("Asset was not found")
		}

		if finalFee, err = fee.convertToWizardFee(ast); err != nil {
			return
		}

		if finalDelegatedStakingClaimAmount, err = ast.ConvertToUnits(delegatedStakingClaimAmount); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateUpdateDelegateTx(from, nonce, delegatedStakingNewPublicKey, delegatedStakingNewFee, finalDelegatedStakingClaimAmount, data, finalFee, propagateTx, awaitAnswer, awaitBroadcast, false, statusCallback)
}

func (builder *TransactionsBuilder) CreateUpdateDelegateTx(from string, nonce uint64, delegatedStakingNewPublicKey []byte, delegatedStakingNewFee, delegatedStakingClaimAmount uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var plainAcc *plain_account.PlainAccount
	var chainHeight uint64

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))

		plainAccs := plain_accounts.NewPlainAccounts(reader)

		if plainAcc, err = plainAccs.GetPlainAccount(fromWalletAddresses[0].PublicKey, chainHeight); err != nil {
			return
		}
		if plainAcc == nil {
			return errors.New("Account doesn't exist")
		}

		return
	}); err != nil {
		return nil, err
	}

	nonce = builder.getNonce(nonce, fromWalletAddresses[0].PublicKey, plainAcc.Nonce)

	if tx, err = wizard.CreateUpdateDelegateTx(nonce, fromWalletAddresses[0].PrivateKey.Key, delegatedStakingNewPublicKey, delegatedStakingNewFee, delegatedStakingClaimAmount, data, fee, false, statusCallback); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func TransactionsBuilderInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain) (builder *TransactionsBuilder) {

	builder = &TransactionsBuilder{
		wallet:  wallet,
		chain:   chain,
		mempool: mempool,
		lock:    &sync.Mutex{},
	}

	builder.initCLI()

	return
}
