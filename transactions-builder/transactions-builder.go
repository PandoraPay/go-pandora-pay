package transactions_builder

import (
	"encoding/binary"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
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

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, nonce uint64, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err2 error) {

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

			available, err := account.GetAvailableBalance(chainHeight, tokens[i])
			if err != nil {
				return err
			}

			if available < amounts[i] {
				return errors.New("Not enough funds")
			}
			if i == 0 && nonce == 0 {
				nonce = builder.mempool.GetNonce(fromWalletAddress.PublicKeyHash, account.Nonce)
			}
			keys[i] = fromWalletAddress.PrivateKey.Key
		}

		if tx, err = wizard.CreateSimpleTx(nonce, keys, amounts, tokens, dsts, dstsAmounts, dstsTokens, feePerByte, feeToken); err != nil {
			return
		}
		for i, fromAddress := range from {
			var fromWalletAddress *wallet.WalletAddress
			if fromWalletAddress, err = builder.wallet.GetWalletAddressByAddress(fromAddress); err != nil {
				return
			}

			account := accs.GetAccountEvenEmpty(fromWalletAddress.PublicKeyHash)
			balance, err := account.GetAvailableBalance(chainHeight, tokens[i])
			if err != nil {
				return err
			}

			if balance < tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Amount {
				return errors.New("You don't have enough coins to pay for the fee")
			}
		}
		return

	})
	return

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

func TransactionsBuilderInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain) (builder *TransactionsBuilder) {

	builder = &TransactionsBuilder{
		wallet:  wallet,
		chain:   chain,
		mempool: mempool,
	}

	builder.initTransactionsBuilderCLI()

	return
}
