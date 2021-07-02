package blockchain_types

import "pandora-pay/blockchain/transactions/transaction"

type BlockchainTransactionUpdate struct {
	TxHash   []byte
	Tx       *transaction.Transaction
	Inserted bool
}
