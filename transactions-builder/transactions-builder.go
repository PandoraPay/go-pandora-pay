package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced-connection-types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/transactions-builder/wizard"
	"pandora-pay/wallet"
	wallet_address "pandora-pay/wallet/address"
	"sync"
)

type TransactionsBuilder struct {
	wallet  *wallet.Wallet
	mempool *mempool.Mempool
	chain   *blockchain.Blockchain
	lock    *sync.Mutex //TODO replace sync.Mutex with a snyc.Map in order to optimize the transactions creation
}

func (builder *TransactionsBuilder) checkTx(acc *account.Account, tx *transaction.Transaction, chainHeight uint64) (err error) {

	if acc == nil {
		return errors.New("Account doesn't even exist")
	}

	base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

	availableDelegatedStake, err := acc.NativeExtra.ComputeDelegatedStakeAvailable(chainHeight)
	if err != nil {
		return err
	}

	switch base.TxScript {
	case transaction_simple.SCRIPT_UNSTAKE:
		if err = helpers.SafeUint64Sub(&availableDelegatedStake, tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).Amount); err != nil {
			return errors.New("You don't have enough staked coins to pay for the withdraw")
		}
	}

	if availableDelegatedStake, err = builder.mempool.GetBalance(base.Vin.PublicKey, availableDelegatedStake, config.NATIVE_TOKEN_FULL); err != nil {
		return
	}

	if availableDelegatedStake < base.Fee {
		return errors.New("You don't have enough staked coins to pay for the fee")
	}

	return nil
}

func (builder *TransactionsBuilder) convertFloatAmounts(amounts []float64, tok *token.Token) ([]uint64, error) {

	var err error

	amountsFinal := make([]uint64, len(amounts))
	for i := range amounts {
		if err != nil {
			return nil, err
		}
		if amountsFinal[i], err = tok.ConvertToUnits(amounts[i]); err != nil {
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

func (builder *TransactionsBuilder) CreateZetherTx_Float(from []string, nonce uint64, token []byte, amounts []float64, dsts []string, dstsAmounts []float64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var amountsFinal, dstsAmountsFinal []uint64

	statusCallback("Converting Floats to Numbers")

	finalFee := &wizard.TransactionsWizardFee{}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks, err := tokens.NewTokens(reader)
		if err != nil {
			return
		}

		tok, err := toks.GetToken(token)
		if tok == nil {
			return errors.New("Token was not found")
		}

		if amountsFinal, err = builder.convertFloatAmounts(amounts, tok); err != nil {
			return
		}
		if dstsAmountsFinal, err = builder.convertFloatAmounts(dstsAmounts, tok); err != nil {
			return
		}

		if finalFee, err = fee.convertToWizardFee(tok); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateZetherTx(from, nonce, token, amountsFinal, dsts, dstsAmountsFinal, data, finalFee, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateZetherTx(from []string, nonce uint64, token []byte, amounts []uint64, dsts []string, dstsAmounts []uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	//fromWalletAddresses, err := builder.getWalletAddresses(from)
	//if err != nil {
	//	return nil, err
	//}
	//
	//statusCallback("Wallet Addresses Found")
	//
	//builder.lock.Lock()
	//defer builder.lock.Unlock()
	//
	//var tx *transaction.Transaction
	//keys := make([][]byte, len(from))
	//accountsList := make([]*account.Account, len(from))
	//
	//if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
	//
	//	accs := accounts.NewAccounts(reader)
	//
	//	chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
	//
	//	for i, fromWalletAddress := range fromWalletAddresses {
	//
	//		if accountsList[i], err = accs.GetAccount(fromWalletAddress.PublicKey, chainHeight); err != nil {
	//			return
	//		}
	//
	//		if accountsList[i] == nil {
	//			return errors.New("Account doesn't exist")
	//		}
	//
	//		if accountsList[i].GetAvailableBalance(token) < amounts[i] {
	//			return errors.New("Not enough funds")
	//		}
	//
	//		keys[i] = fromWalletAddress.PrivateKey.Key
	//	}
	//
	//	return
	//}); err != nil {
	//	return nil, err
	//}
	//
	//statusCallback("Balances checked")
	//
	//if nonce == 0 {
	//	nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKey, accountsList[0].Nonce)
	//}
	//
	//statusCallback("Getting Nonce from Mempool")
	//
	//if tx, err = wizard.CreateZetherTx(nonce, token, keys, amounts, dsts, dstsAmounts, data, fee, statusCallback); err != nil {
	//	gui.GUI.Error("Error creating Tx: ", err)
	//	return nil, err
	//}
	//
	//statusCallback("Transaction Created")
	//
	//if err = builder.checkTx(accountsList, tx); err != nil {
	//	return nil, err
	//}
	//
	//statusCallback("Tx checked")
	//
	//if propagateTx {
	//	if err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
	//		return nil, err
	//	}
	//}
	//
	//return tx, nil

	return nil, errors.New("Skip")
}

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(from string, nonce uint64, unstakeAmount float64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	statusCallback("Converting Floats to Numbers")

	unstakeAmountFinal, err := config.ConvertToUnits(unstakeAmount)
	if err != nil {
		return nil, err
	}

	feeFinal := &wizard.TransactionsWizardFee{}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks, err := tokens.NewTokens(reader)
		if err != nil {
			return
		}

		tok, err := toks.GetToken(config.NATIVE_TOKEN_FULL)
		if err != nil {
			return
		}
		if tok == nil {
			return errors.New("Token was not found")
		}

		if feeFinal, err = fee.convertToWizardFee(tok); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateUnstakeTx(from, nonce, unstakeAmountFinal, data, feeFinal, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, nonce, unstakeAmount uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	statusCallback("Wallet Addresses Found")

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var acc *account.Account

	var chainHeight uint64
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		accsCollection := accounts.NewAccountsCollection(reader)

		accs, err := accsCollection.GetMap(config.NATIVE_TOKEN_FULL)
		if err != nil {
			return
		}

		if acc, err = accs.GetAccount(fromWalletAddresses[0].PublicKey); err != nil {
			return
		}
		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		if err = acc.NativeExtra.RefreshDelegatedStake(chainHeight); err != nil {
			return
		}

		availableStake, err := acc.NativeExtra.ComputeDelegatedStakeAvailable(chainHeight)
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

	if nonce == 0 {
		nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKey, acc.NativeExtra.Nonce)
	}
	statusCallback("Getting Nonce from Mempool")

	if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddresses[0].PrivateKey.Key, unstakeAmount, data, fee, statusCallback); err != nil {
		return nil, err
	}
	statusCallback("Transaction Created")

	if err = builder.checkTx(acc, tx, chainHeight); err != nil {
		return nil, err
	}

	statusCallback("Tx checked")

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TransactionsBuilder) CreateUpdateDelegateTx_Float(from string, nonce uint64, delegateNewPubKeyGenerate bool, delegateNewPubKey []byte, delegateNewFee uint64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var finalFee *wizard.TransactionsWizardFee

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks, err := tokens.NewTokens(reader)
		if err != nil {
			return
		}

		tok, err := toks.GetToken(config.NATIVE_TOKEN_FULL)
		if err != nil {
			return
		}
		if tok == nil {
			return errors.New("Token was not found")
		}

		if finalFee, err = fee.convertToWizardFee(tok); err != nil {
			return err
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateUpdateDelegateTx(from, nonce, delegateNewPubKeyGenerate, delegateNewPubKey, delegateNewFee, data, finalFee, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateUpdateDelegateTx(from string, nonce uint64, delegateNewPubKeyGenerate bool, delegateNewPubKey []byte, delegateNewFee uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var acc *account.Account
	var chainHeight uint64

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))

		accsCollection := accounts.NewAccountsCollection(reader)

		accs, err := accsCollection.GetMap(config.NATIVE_TOKEN_FULL)
		if err != nil {
			return
		}

		if acc, err = accs.GetAccount(fromWalletAddresses[0].PublicKey); err != nil {
			return
		}
		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		if err = acc.NativeExtra.RefreshDelegatedStake(chainHeight); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	if nonce == 0 {
		nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKey, acc.NativeExtra.Nonce)
	}

	if delegateNewPubKeyGenerate {

		var delegatedStake *wallet_address.WalletAddressDelegatedStake
		if delegatedStake, err = fromWalletAddresses[0].DeriveDelegatedStake(uint32(nonce)); err != nil {
			return nil, err
		}
		delegateNewPubKey = delegatedStake.PublicKey

	}

	if tx, err = wizard.CreateUpdateDelegateTx(nonce, fromWalletAddresses[0].PrivateKey.Key, delegateNewPubKey, delegateNewFee, data, fee, statusCallback); err != nil {
		return nil, err
	}

	if err = builder.checkTx(acc, tx, chainHeight); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
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
