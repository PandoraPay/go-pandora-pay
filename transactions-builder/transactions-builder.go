package transactions_builder

import (
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/wizard"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/wallet"
)

type TransactionsBuilder struct {
	wallet *wallet.Wallet
	chain  *blockchain.Blockchain
}

func (builder *TransactionsBuilder) CreateSimpleTx(from []string, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {
	return
}

func (builder *TransactionsBuilder) CreateUnstakeTx(from string, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction, err error) {

	defer func() {
		if err2 := recover(); err2 != nil {
			err = helpers.ConvertRecoverError(err2)
		}
	}()

	fromWalletAddress := builder.wallet.GetWalletAddressByAddress(from)

	if err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {

		accs := accounts.NewAccounts(boltTx)
		account := accs.GetAccount(fromWalletAddress.PublicKeyHash)
		if account == nil {
			panic("Account doesn't exist")
		}

		tx = wizard.CreateUnstakeTx(account.Nonce, fromWalletAddress.PrivateKey.Key, unstakeAmount, feePerByte, feeToken, payFeeInExtra)

		return nil
	}); err != nil {
		panic(err)
	}

	return
}

func TransactionsBuilderInit(wallet *wallet.Wallet, chain *blockchain.Blockchain) (builder *TransactionsBuilder) {

	builder = &TransactionsBuilder{
		wallet: wallet,
		chain:  chain,
	}

	return
}
