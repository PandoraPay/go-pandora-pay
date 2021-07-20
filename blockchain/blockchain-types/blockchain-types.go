package blockchain_types

import "pandora-pay/blockchain/transactions/transaction"

type BlockchainTransactionKeyUpdate struct {
	PublicKeyHash []byte
	TxsCount      uint64
}

type BlockchainTransactionUpdate struct {
	TxHash      []byte
	Tx          *transaction.Transaction
	Inserted    bool
	BlockHeight uint64
	Keys        []*BlockchainTransactionKeyUpdate
}
