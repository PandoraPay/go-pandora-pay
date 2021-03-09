package transactions_builder

import (
	"encoding/binary"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/wizard"
	"pandora-pay/store"
	"pandora-pay/wallet"
)

type TransactionsBuilder struct {
	wallet *wallet.Wallet
	chain  *blockchain.Blockchain
}

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction) {

	if err := store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {
		reader := boltTx.Bucket([]byte("Chain"))
		accs := accounts.NewAccounts(boltTx)

		buffer := reader.Get([]byte("chainHeight"))
		chainHeight, _ := binary.Uvarint(buffer)

		var nonce uint64
		var keys [][32]byte
		for i, fromAddress := range from {
			fromWalletAddress := builder.wallet.GetWalletAddressByAddress(fromAddress)
			account := accs.GetAccountEvenEmpty(fromWalletAddress.PublicKeyHash)
			if account == nil {
				panic("Account doesn't exist")
			}

			available := account.GetAvailableBalance(chainHeight, tokens[i])
			if available < amounts[i] {
				panic("Not enough funds")
			}
			if i == 0 {
				nonce = account.Nonce
			}
			keys = append(keys, fromWalletAddress.PrivateKey.Key)
		}

		tx = wizard.CreateSimpleTx(nonce, keys, amounts, tokens, dsts, dstsAmounts, dstsTokens, feePerByte, feeToken)
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

		tx = wizard.CreateUnstakeTx(account.Nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra)
		return nil

	}); err != nil {
		panic(err)
	}

	return
}

func TransactionsBuilderInit(wallet *wallet.Wallet, chain *blockchain.Blockchain) *TransactionsBuilder {
	return &TransactionsBuilder{
		wallet: wallet,
		chain:  chain,
	}
}
