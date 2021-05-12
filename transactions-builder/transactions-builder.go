package transactions_builder

import (
	"encoding/binary"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/store"
	"pandora-pay/transactions-builder/wizard"
	"pandora-pay/wallet"
	wallet_address "pandora-pay/wallet/address"
)

type TransactionsBuilder struct {
	wallet  *wallet.Wallet
	mempool *mempool.Mempool
	chain   *blockchain.Blockchain
}

func (builder *TransactionsBuilder) checkTx(accs *accounts.Accounts, chainHeight uint64, tx *transaction.Transaction) (err error) {

	var available uint64
	for _, vin := range tx.TxBase.(*transaction_simple.TransactionSimple).Vin {

		var acc *account.Account

		acc, err = accs.GetAccount(vin.Bloom.PublicKeyHash, chainHeight)
		if err != nil {
			return
		}

		if acc == nil {
			return errors.New("Account doesn't even exist")
		}

		available, err = acc.GetAvailableBalance(vin.Token)
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

			var tok *token.Token
			if tok, err = toks.GetToken(amountsTokens[i]); err != nil {
				return
			}

			if tok == nil {
				return errors.New("Token was not found")
			}
			if amountsFinal[i], err = tok.ConvertToUnits(amounts[i]); err != nil {
				return
			}
		}
		for i := range dstsTokens {
			var tok *token.Token
			if tok, err = toks.GetToken(dstsTokens[i]); err != nil {
				return
			}

			if tok == nil {
				return errors.New("Token was not found")
			}
			if dstsAmountsFinal[i], err = tok.ConvertToUnits(dstsAmounts[i]); err != nil {
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

		accs := accounts.NewAccounts(boltTx)

		chainHeight, _ := binary.Uvarint(boltTx.Bucket([]byte("Chain")).Get([]byte("chainHeight")))

		keys := make([][]byte, len(from))

		for i, fromAddress := range from {

			var fromWalletAddress *wallet_address.WalletAddress
			if fromWalletAddress, err = builder.wallet.GetWalletAddressByAddress(fromAddress); err != nil {
				return
			}

			var acc *account.Account
			if acc, err = accs.GetAccount(fromWalletAddress.GetPublicKeyHash(), chainHeight); err != nil {
				return
			}

			if acc == nil {
				return errors.New("Account doesn't exist")
			}

			var available uint64
			if available, err = acc.GetAvailableBalance(amountsTokens[i]); err != nil {
				return err
			}
			if available < amounts[i] {
				return errors.New("Not enough funds")
			}

			if i == 0 && nonce == 0 {
				nonce = builder.mempool.GetNonce(fromWalletAddress.GetPublicKeyHash(), acc.Nonce)
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

		chainHeight, _ := binary.Uvarint(boltTx.Bucket([]byte("Chain")).Get([]byte("chainHeight")))

		accs := accounts.NewAccounts(boltTx)
		var account *account.Account
		if account, err = accs.GetAccount(fromWalletAddress.GetPublicKeyHash(), chainHeight); err != nil {
			return
		}

		if account == nil {
			return errors.New("Account doesn't exist")
		}

		availableUnstake := account.GetDelegatedStakeAvailable()

		if availableUnstake < unstakeAmount {
			return errors.New("You don't have enough staked coins")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.GetPublicKeyHash(), account.Nonce)
		}

		if tx, err = wizard.CreateUnstakeTx(nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra); err != nil {
			return
		}

		availableDelegatedStake := account.GetDelegatedStakeAvailable()

		if availableDelegatedStake < tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Amount+tx.TxBase.(*transaction_simple.TransactionSimple).Extra.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra {
			return errors.New("You don't have enough staked coins to pay for the fee")
		}

		return
	})

	return
}

func (builder *TransactionsBuilder) CreateDelegateTx_Float(from string, nonce uint64, delegateAmount float64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {

	delegateAmountFinal, err := config.ConvertToUnits(delegateAmount)
	if err != nil {
		return
	}

	return builder.CreateDelegateTx(from, nonce, delegateAmountFinal, delegateNewPubKeyHashGenerate, delegateNewPubKeyHash, feePerByte, feeToken)
}

func (builder *TransactionsBuilder) CreateDelegateTx(from string, nonce uint64, delegateAmount uint64, delegateNewPubKeyHashGenerate bool, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err2 error) {

	fromWalletAddress, err2 := builder.wallet.GetWalletAddressByAddress(from)
	if err2 != nil {
		return
	}

	err2 = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		chainHeight, _ := binary.Uvarint(boltTx.Bucket([]byte("Chain")).Get([]byte("chainHeight")))

		accs := accounts.NewAccounts(boltTx)
		var acc *account.Account
		if acc, err = accs.GetAccount(fromWalletAddress.GetPublicKeyHash(), chainHeight); err != nil {
			return
		}

		if acc == nil {
			return errors.New("Account doesn't exist")
		}

		var available uint64
		if available, err = acc.GetAvailableBalance(config.NATIVE_TOKEN); err != nil {
			return
		}

		if available < delegateAmount {
			return errors.New("You don't have enough coins to delegate")
		}

		if nonce == 0 {
			nonce = builder.mempool.GetNonce(fromWalletAddress.GetPublicKeyHash(), acc.Nonce)
		}

		if delegateNewPubKeyHashGenerate {

			var delegatedStake *wallet_address.WalletAddressDelegatedStake

			if delegatedStake, err = fromWalletAddress.DeriveDelegatedStake(uint32(nonce)); err != nil {
				return
			}
			delegateNewPubKeyHash = delegatedStake.PublicKeyHash

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

	builder.initCLI()

	return
}
