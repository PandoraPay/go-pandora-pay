package transactions_builder

import (
	"encoding/binary"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/store"
	"pandora-pay/transactions-builder/wizard"
	"pandora-pay/wallet"
)

type TransactionsBuilder struct {
	wallet  *wallet.Wallet
	mempool *mempool.Mempool
	chain   *blockchain.Blockchain
}

func (builder *TransactionsBuilder) checkTx(accs *accounts.Accounts, chainHeight uint64, tx *transaction.Transaction) (err error) {

	var available uint64
	for _, vin := range tx.TxBase.(*transaction_simple.TransactionSimple).Vin {
		account := accs.GetAccountEvenEmpty(vin.Bloom.PublicKeyHash)
		available, err = account.GetAvailableBalance(chainHeight, vin.Token)
		if err != nil {
			return err
		}

		if available, err = builder.mempool.GetBalance(vin.Bloom.PublicKeyHash, available, vin.Token); err != nil {
			return
		}
		if available < vin.Amount {
			return errors.New("You don't have enough coins")
		}
	}
	return
}

func (builder *TransactionsBuilder) CreateSimpleTx_Float(from []string, nonce uint64, amounts []float64, amountsTokens [][]byte, dsts []string, dstsAmounts []float64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err2 error) {

	amountsFinal := make([]uint64, len(from))
	dstsAmountsFinal := make([]uint64, len(dsts))

	if err2 = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		toks := tokens.NewTokens(boltTx)
		for i := range from {
			token := toks.GetToken(amountsTokens[i])
			if token == nil {
				return errors.New("Token was not found")
			}
			if amountsFinal[i], err = token.ConvertToUnits(amounts[i]); err != nil {
				return
			}
		}
		for i := range dstsTokens {
			token := toks.GetToken(dstsTokens[i])
			if token == nil {
				return errors.New("Token was not found")
			}
			if dstsAmountsFinal[i], err = token.ConvertToUnits(dstsAmounts[i]); err != nil {
				return
			}
		}

		return
	}); err2 != nil {
		return
	}

	return builder.CreateSimpleTx(from, nonce, amountsFinal, amountsTokens, dsts, dstsAmountsFinal, dstsTokens, feePerByte, feeToken)
}

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, nonce uint64, amounts []uint64, amountsTokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err2 error) {

	err2 = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		reader := boltTx.Bucket([]byte("Chain"))
		accs := accounts.NewAccounts(boltTx)

		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		keys := make([][]byte, len(from))

		for i, fromAddress := range from {

			var fromWalletAddress *wallet.WalletAddress
			if fromWalletAddress, err = builder.wallet.GetWalletAddressByAddress(fromAddress); err != nil {
				return
			}

			account := accs.GetAccount(fromWalletAddress.PublicKeyHash)
			if account == nil {
				return errors.New("Account doesn't exist")
			}

			var available uint64
			if available, err = account.GetAvailableBalance(chainHeight, amountsTokens[i]); err != nil {
			}
			if available < amounts[i] {
				return errors.New("Not enough funds")
			}

			if i == 0 && nonce == 0 {
				nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, account.Nonce)
			}
			keys[i] = fromWalletAddress.PrivateKey.Key
		}

		if tx, err = wizard.CreateSimpleTx(nonce, keys, amounts, amountsTokens, dsts, dstsAmounts, dstsTokens, feePerByte, feeToken); err != nil {
			return
		}

		if err = builder.checkTx(accs, chainHeight, tx); err != nil {
			return
		}

		return
	})
	return

}

func (builder *TransactionsBuilder) CreateUnstakeTx_Float(from string, nonce uint64, unstakeAmount float64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction, err2 error) {

	unstakeAmountFinal, err2 := config.ConvertToUnits(unstakeAmount)
	if err2 != nil {
		return
	}

	return builder.CreateUnstakeTx(from, nonce, unstakeAmountFinal, feePerByte, feeToken, payFeeInExtra)
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, nonce uint64, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction, err2 error) {

	fromWalletAddress, err2 := builder.wallet.GetWalletAddressByAddress(from)
	if err2 != nil {
		return
	}

	err2 = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		account := accounts.NewAccounts(boltTx).GetAccount(fromWalletAddress.PublicKeyHash)
		if account == nil {
			return errors.New("Account doesn't exist")
		}

		availableUnstake, err := account.GetDelegatedStakeAvailable(chainHeight)
		if err != nil {
			return
		}
		if availableUnstake < unstakeAmount {
			return errors.New("You don't have enough staked coins")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, account.Nonce)
		}

		if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra); err != nil {
			return
		}

		var availableDelegatedStake uint64
		if availableDelegatedStake, err = account.GetDelegatedStakeAvailable(chainHeight); err != nil {
			return err
		}
		if availableDelegatedStake < tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Amount+tx.TxBase.(*transaction_simple.TransactionSimple).Extra.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra {
			return errors.New("You don't have enough staked coins to pay for the fee")
		}

		return
	})

	return
}

func (builder *TransactionsBuilder) CreateDelegateTx_Float(from string, nonce uint64, delegateAmount float64, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {

	delegateAmountFinal, err := config.ConvertToUnits(delegateAmount)
	if err != nil {
		return
	}

	return builder.CreateDelegateTx(from, nonce, delegateAmountFinal, delegateNewPubKeyHash, feePerByte, feeToken)
}

func (builder *TransactionsBuilder) CreateDelegateTx(from string, nonce uint64, delegateAmount uint64, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err2 error) {

	fromWalletAddress, err2 := builder.wallet.GetWalletAddressByAddress(from)
	if err2 != nil {
		return
	}

	err2 = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Chain"))
		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		accs := accounts.NewAccounts(boltTx)
		account := accs.GetAccount(fromWalletAddress.PublicKeyHash)
		if account == nil {
			return errors.New("Account doesn't exist")
		}

		available, err := account.GetAvailableBalance(chainHeight, config.NATIVE_TOKEN)
		if err != nil {
			return
		}
		if available < delegateAmount {
			return errors.New("You don't have enough coins to delegate")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, account.Nonce)
		}

		if tx, err = wizard.CreateDelegateTx(nonce, fromWalletAddress.PrivateKey.Key, delegateAmount, delegateNewPubKeyHash, feePerByte, feeToken); err != nil {
			return
		}

		if err = builder.checkTx(accs, chainHeight, tx); err != nil {
			return
		}

		return
	})

	return
}

func TransactionsBuilderInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain) (builder *TransactionsBuilder) {

	builder = &TransactionsBuilder{
		wallet:  wallet,
		chain:   chain,
		mempool: mempool,
	}

	builder.initTransactionsBuilderCLI()

	return
}
