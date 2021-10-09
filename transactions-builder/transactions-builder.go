package transactions_builder

import (
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	plain_accounts "pandora-pay/blockchain/data_storage/plain-accounts"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	"pandora-pay/blockchain/data_storage/registrations"
	"pandora-pay/blockchain/data_storage/tokens"
	"pandora-pay/blockchain/data_storage/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	transaction_simple_parts "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	"pandora-pay/config"
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

func (builder *TransactionsBuilder) getNonce(nonce uint64, publicKey []byte, accNonce uint64) uint64 {
	if nonce != 0 {
		return nonce
	}
	return builder.mempool.GetNonce(publicKey, accNonce)
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

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(from string, nonce uint64, unstakeAmount float64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(status string)) (*transaction.Transaction, error) {

	statusCallback("Converting Floats to Numbers")

	unstakeAmountFinal, err := config.ConvertToUnits(unstakeAmount)
	if err != nil {
		return nil, err
	}

	feeFinal := &wizard.TransactionsWizardFee{}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)

		tok, err := toks.GetToken(config.NATIVE_TOKEN)
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

func (builder *TransactionsBuilder) CreateUpdateDelegateTx_Float(from string, nonce uint64, delegateNewPubKeyGenerate bool, delegateNewPubKey []byte, delegateNewFee uint64, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var finalFee *wizard.TransactionsWizardFee

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)

		tok, err := toks.GetToken(config.NATIVE_TOKEN)
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

	return builder.CreateUpdateDelegateTx(from, nonce, delegateNewPubKeyGenerate, delegateNewPubKey, delegateNewFee, data, finalFee, propagateTx, awaitAnswer, awaitBroadcast, false, statusCallback)
}

func (builder *TransactionsBuilder) CreateUpdateDelegateTx(from string, nonce uint64, delegateNewPubKeyGenerate bool, delegateNewPubKey []byte, delegateNewFee uint64, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

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

	if delegateNewPubKeyGenerate {
		var delegatedStake *wallet_address.WalletAddressDelegatedStake
		if delegatedStake, err = fromWalletAddresses[0].DeriveDelegatedStake(uint32(nonce)); err != nil {
			return nil, err
		}
		delegateNewPubKey = delegatedStake.PublicKey
	}

	if tx, err = wizard.CreateUpdateDelegateTx(nonce, fromWalletAddresses[0].PrivateKey.Key, delegateNewPubKey, delegateNewFee, data, fee, false, statusCallback); err != nil {
		return nil, err
	}

	if propagateTx {
		if err = builder.mempool.AddTxToMemPool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func (builder *TransactionsBuilder) CreateClaimTx_Float(from string, nonce uint64, outputAmounts []float64, outputAddresses []string, data *wizard.TransactionsWizardData, fee *TransactionsBuilderFeeFloat, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

	var finalFee *wizard.TransactionsWizardFee
	outputAmountsFinal := make([]uint64, len(outputAmounts))

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		toks := tokens.NewTokens(reader)

		tok, err := toks.GetToken(config.NATIVE_TOKEN)
		if err != nil {
			return
		}
		if tok == nil {
			return errors.New("Token was not found")
		}

		if finalFee, err = fee.convertToWizardFee(tok); err != nil {
			return err
		}

		for i, amount := range outputAmounts {
			if outputAmountsFinal[i], err = tok.ConvertToUnits(amount); err != nil {
				return
			}
		}

		return
	}); err != nil {
		return nil, err
	}

	return builder.CreateClaimTx(from, nonce, outputAmountsFinal, outputAddresses, data, finalFee, propagateTx, awaitAnswer, awaitBroadcast, validateTx, statusCallback)
}

func (builder *TransactionsBuilder) CreateClaimTx(from string, nonce uint64, outputAmounts []uint64, outputAddresses []string, data *wizard.TransactionsWizardData, fee *wizard.TransactionsWizardFee, propagateTx, awaitAnswer, awaitBroadcast bool, validateTx bool, statusCallback func(string)) (*transaction.Transaction, error) {

	fromWalletAddresses, err := builder.getWalletAddresses([]string{from})
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	var tx *transaction.Transaction
	var plainAcc *plain_account.PlainAccount
	var chainHeight uint64

	output := make([]*transaction_simple_parts.TransactionSimpleOutput, len(outputAmounts))
	txRegistrations := make([]*transaction_data.TransactionDataRegistration, 0)

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))

		plainAccs := plain_accounts.NewPlainAccounts(reader)
		regs := registrations.NewRegistrations(reader)

		if plainAcc, err = plainAccs.GetPlainAccount(fromWalletAddresses[0].PublicKey, chainHeight); err != nil {
			return
		}
		if plainAcc == nil {
			return errors.New("Account doesn't exist")
		}

		for i := range outputAmounts {

			var addr *addresses.Address
			if addr, err = addresses.DecodeAddr(outputAddresses[i]); err != nil {
				return
			}

			var isReg bool
			if isReg, err = regs.Exists(string(addr.PublicKey)); err != nil {
				return
			}

			output[i] = &transaction_simple_parts.TransactionSimpleOutput{
				Amount:    outputAmounts[i],
				PublicKey: addr.PublicKey,
			}

			if !isReg {
				if addr.Registration == nil {
					return errors.New("Registration is missing one of the specified addresses")
				}

				txRegistrations = append(txRegistrations, &transaction_data.TransactionDataRegistration{
					PublicKeyIndex:        uint64(i),
					RegistrationSignature: addr.Registration,
				})

			}

		}

		return
	}); err != nil {
		return nil, err
	}

	nonce = builder.getNonce(nonce, fromWalletAddresses[0].PublicKey, plainAcc.Nonce)

	if tx, err = wizard.CreateClaimTx(nonce, fromWalletAddresses[0].PrivateKey.Key, txRegistrations, output, data, fee, validateTx, statusCallback); err != nil {
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
