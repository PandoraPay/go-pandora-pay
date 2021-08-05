package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
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

func (builder *TransactionsBuilder) checkTx(accountsList []*account.Account, tx *transaction.Transaction) (err error) {

	var available uint64
	for i, vin := range tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin {

		if accountsList[i] == nil {
			return errors.New("Account doesn't even exist")
		}

		available = accountsList[i].GetAvailableBalance(vin.Token)

		if available, err = builder.mempool.GetBalance(vin.Bloom.PublicKeyHash, available, vin.Token); err != nil {
			return
		}
		if available < vin.Amount {
			return errors.New("You don't have enough coins")
		}
	}
	return
}

func (builder *TransactionsBuilder) convertFloatAmounts(amounts []float64, tokens [][]byte, toks *tokens.Tokens) ([]uint64, error) {

	if len(tokens) != len(amounts) {
		return nil, errors.New("Amounts len is not matching tokens len")
	}

	amountsFinal := make([]uint64, len(amounts))
	for i := range amounts {
		tok, err := toks.GetTokenRequired(tokens[i])
		if err != nil {
			return nil, err
		}
		if amountsFinal[i], err = tok.ConvertToUnits(amounts[i]); err != nil {
			return nil, err
		}
	}

	return amountsFinal, nil
}

func (builder *TransactionsBuilder) convertFloatFees(feeFixed, feePerByte float64, feeToken []byte, toks *tokens.Tokens) (feeFixedFinal, feePerByteFinal uint64, err error) {

	tok, err := toks.GetTokenRequired(feeToken)
	if err != nil {
		return
	}

	if feeFixed > 0 {
		if feeFixedFinal, err = tok.ConvertToUnits(feeFixed); err != nil {
			return
		}
	}
	if feePerByte > 0 {
		if feePerByteFinal, err = tok.ConvertToUnits(feePerByte); err != nil {
			return
		}
	}

	return
}

func (builder *TransactionsBuilder) getWalletAddresses(from []string) ([]*wallet_address.WalletAddress, error) {

	fromWalletAddress := make([]*wallet_address.WalletAddress, len(from))
	var err error

	for i, fromAddress := range from {
		if fromWalletAddress[i], err = builder.wallet.GetWalletAddressByEncodedAddress(fromAddress); err != nil {
			return nil, err
		}
	}

	return fromWalletAddress, nil
}

func (builder *TransactionsBuilder) CreateSimpleTx_Float(from []string, nonce uint64, amounts []float64, amountsTokens [][]byte, dsts []string, dstsAmounts []float64, dstsTokens [][]byte, feeFixed, feePerByte float64, feePerByteAuto bool, feeToken []byte, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var feeFixedFinal, feePerByteFinal uint64
	var amountsFinal, dstsAmountsFinal []uint64

	statusCallback("Converting Floats to Numbers")

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)

		if amountsFinal, err = builder.convertFloatAmounts(amounts, amountsTokens, toks); err != nil {
			return
		}
		if dstsAmountsFinal, err = builder.convertFloatAmounts(dstsAmounts, dstsTokens, toks); err != nil {
			return
		}

		if feeFixedFinal, feePerByteFinal, err = builder.convertFloatFees(feeFixed, feePerByte, feeToken, toks); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateSimpleTx(from, nonce, amountsFinal, amountsTokens, dsts, dstsAmountsFinal, dstsTokens, feeFixedFinal, feePerByteFinal, feePerByteAuto, feeToken, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, nonce uint64, amounts []uint64, amountsTokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses(from)
	if err != nil {
		return nil, err
	}

	statusCallback("Wallet Addresses Found")

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	keys := make([][]byte, len(from))
	accountsList := make([]*account.Account, len(from))

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accs := accounts.NewAccounts(reader)

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		for i, fromWalletAddress := range fromWalletAddresses {

			if accountsList[i], err = accs.GetAccount(fromWalletAddress.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if accountsList[i] == nil {
				return errors.New("Account doesn't exist")
			}

			if accountsList[i].GetAvailableBalance(amountsTokens[i]) < amounts[i] {
				return errors.New("Not enough funds")
			}

			keys[i] = fromWalletAddress.PrivateKey.Key
		}

		return
	}); err != nil {
		return nil, err
	}

	statusCallback("Balances checked")

	if nonce == 0 {
		nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKeyHash, accountsList[0].Nonce)
	}

	statusCallback("Getting Nonce from Mempool")

	if tx, err = wizard.CreateSimpleTx(nonce, keys, amounts, amountsTokens, dsts, dstsAmounts, dstsTokens, feeFixed, feePerByte, feePerByteAuto, feeToken, statusCallback); err != nil {
		gui.GUI.Error("Error creating Tx: ", err)
		return nil, err
	}

	statusCallback("Transaction Created")

	if err = builder.checkTx(accountsList, tx); err != nil {
		return nil, err
	}

	statusCallback("Tx checked")

	if propagateTx {
		if err := builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, ""); err != nil {
			return nil, err
		}
	}

	return tx, nil

}

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(from string, nonce uint64, unstakeAmount float64, feeFixed, feePerByte float64, feePerByteAuto bool, feeToken []byte, feePayInExtra, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	statusCallback("Converting Floats to Numbers")

	unstakeAmountFinal, err := config.ConvertToUnits(unstakeAmount)
	if err != nil {
		return nil, err
	}

	var feeFixedFinal, feePerByteFinal uint64

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)
		if feeFixedFinal, feePerByteFinal, err = builder.convertFloatFees(feeFixed, feePerByte, feeToken, toks); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateUnstakeTx(from, nonce, unstakeAmountFinal, feeFixedFinal, feePerByteFinal, feePerByteAuto, feeToken, feePayInExtra, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, nonce uint64, unstakeAmount, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, feePayInExtra, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	statusCallback("Wallet Addresses Found")

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	accountsList := make([]*account.Account, 1)

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)

		if accountsList[0], err = accs.GetAccount(fromWalletAddresses[0].PublicKeyHash, chainHeight); err != nil {
			return
		}

		if accountsList[0] == nil {
			return errors.New("Account doesn't exist")
		}

		if accountsList[0].GetDelegatedStakeAvailable() < unstakeAmount {
			return errors.New("You don't have enough staked coins")
		}

		return
	}); err != nil {
		return nil, err
	}

	statusCallback("Balances checked")

	if nonce == 0 {
		nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKeyHash, accountsList[0].Nonce)
	}
	statusCallback("Getting Nonce from Mempool")

	if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddresses[0].PrivateKey.Key, unstakeAmount, feeFixed, feePerByte, feePerByteAuto, feeToken, feePayInExtra, statusCallback); err != nil {
		return nil, err
	}
	statusCallback("Transaction Created")

	if err = builder.checkTx(accountsList, tx); err != nil {
		return nil, err
	}

	availableDelegatedStake := accountsList[0].GetDelegatedStakeAvailable()
	if availableDelegatedStake < tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin[0].Amount+tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra {
		return nil, errors.New("You don't have enough staked coins to pay for the fee")
	}

	statusCallback("Tx checked")

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, ""); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TransactionsBuilder) CreateDelegateTx_Float(from string, nonce uint64, delegateAmount float64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feeFixed, feePerByte float64, feePerByteAuto bool, feeToken []byte, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	delegateAmountFinal, err := config.ConvertToUnits(delegateAmount)
	if err != nil {
		return nil, err
	}

	var feeFixedFinal, feePerByteFinal uint64

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)
		if feeFixedFinal, feePerByteFinal, err = builder.convertFloatFees(feeFixed, feePerByte, feeToken, toks); err != nil {
			return
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateDelegateTx(from, nonce, delegateAmountFinal, delegateNewPubKeyHashGenerate, delegateNewPubKeyHash, feeFixedFinal, feePerByteFinal, feePerByteAuto, feeToken, propagateTx, awaitAnswer, awaitBroadcast, statusCallback)
}

func (builder *TransactionsBuilder) CreateDelegateTx(from string, nonce uint64, delegateAmount uint64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, propagateTx, awaitAnswer, awaitBroadcast bool, statusCallback func(string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	accountsList := make([]*account.Account, 1)
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)

		if accountsList[0], err = accs.GetAccount(fromWalletAddresses[0].PublicKeyHash, chainHeight); err != nil {
			return
		}

		if accountsList[0] == nil {
			return errors.New("Account doesn't exist")
		}

		available := accountsList[0].GetAvailableBalance(config.NATIVE_TOKEN)
		if available < delegateAmount {
			return errors.New("You don't have enough coins to delegate")
		}

		return
	}); err != nil {
		return nil, err
	}

	if nonce == 0 {
		nonce = builder.mempool.GetNonce(fromWalletAddresses[0].PublicKeyHash, accountsList[0].Nonce)
	}

	if delegateNewPubKeyHashGenerate {

		var delegatedStake *wallet_address.WalletAddressDelegatedStake
		if delegatedStake, err = fromWalletAddresses[0].DeriveDelegatedStake(uint32(nonce)); err != nil {
			return nil, err
		}
		delegateNewPubKeyHash = delegatedStake.PublicKeyHash

	}

	if tx, err = wizard.CreateDelegateTx(nonce, fromWalletAddresses[0].PrivateKey.Key, delegateAmount, delegateNewPubKeyHash, feeFixed, feePerByte, feePerByteAuto, feeToken, statusCallback); err != nil {
		return nil, err
	}

	if err = builder.checkTx(accountsList, tx); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, builder.chain.GetChainData().Height, awaitAnswer, awaitBroadcast, ""); err != nil {
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
