package transactions_builder

import (
	"encoding/binary"
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
	memPool *mempool.MemPool
	chain   *blockchain.Blockchain
}

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction) {

	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		accs := accounts.NewAccounts(boltTx)

		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		var nonce uint64
		var keys [][]byte
		for i, fromAddress := range from {
			fromWalletAddress := builder.wallet.GetWalletAddressByAddress(fromAddress)
			account := accs.GetAccount(fromWalletAddress.PublicKeyHash)
			if account == nil {
				panic("Account doesn't exist")
			}

			available := account.GetAvailableBalance(chainHeight, tokens[i])
			if available < amounts[i] {
				panic("Not enough funds")
			}
			if i == 0 {
				var result bool
				result, nonce = builder.memPool.GetNonce(fromWalletAddress.PublicKeyHash)
				if !result {
					nonce = account.Nonce
				}
			}
			keys = append(keys, fromWalletAddress.PrivateKey.Key)
		}

		tx = wizard.CreateSimpleTx(nonce, keys, amounts, tokens, dsts, dstsAmounts, dstsTokens, feePerByte, feeToken)
		for i, fromAddress := range from {
			fromWalletAddress := builder.wallet.GetWalletAddressByAddress(fromAddress)
			account := accs.GetAccountEvenEmpty(fromWalletAddress.PublicKeyHash)
			if account.GetAvailableBalance(chainHeight, tokens[i]) < tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Amount {
				panic("You don't have enough coins to pay for the fee")
			}
		}
		return nil

	}); err != nil {
		panic(err)
	}

	return
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction) {
	fromWalletAddress := builder.wallet.GetWalletAddressByAddress(from)

	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		account := accounts.NewAccounts(boltTx).GetAccount(fromWalletAddress.PublicKeyHash)
		if account == nil {
			panic("Account doesn't exist")
		}
		if account.GetDelegatedStakeAvailable(chainHeight) < unstakeAmount {
			panic("You don't have enough staked coins")
		}

		result, nonce := builder.memPool.GetNonce(fromWalletAddress.PublicKeyHash)
		if !result {
			nonce = account.Nonce
		}

		tx = wizard.CreateUnstakeTx(nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra)
		if account.GetDelegatedStakeAvailable(chainHeight) < tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Amount+tx.TxBase.(*transaction_simple.TransactionSimple).Extra.(*transaction_simple_extra.TransactionSimpleUnstake).UnstakeFeeExtra {
			panic("You don't have enough staked coins to pay for the fee")
		}

		return nil

	}); err != nil {
		panic(err)
	}

	return
}

func TransactionsBuilderInit(wallet *wallet.Wallet, memPool *mempool.MemPool, chain *blockchain.Blockchain) *TransactionsBuilder {
	return &TransactionsBuilder{
		wallet:  wallet,
		chain:   chain,
		memPool: memPool,
	}
}
