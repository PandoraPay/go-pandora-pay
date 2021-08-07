package blockchain_types

import "pandora-pay/blockchain/transactions/transaction"

type BlockchainTransactionKeyUpdate struct {
	PublicKeyHash []byte
	TxsCount      uint64
}

type BlockchainTransactionUpdate struct {
	TxHash         []byte
	TxHashStr      string
	Tx             *transaction.Transaction
	Inserted       bool
	BlockHeight    uint64
	BlockTimestamp uint64
	Height         uint64
	Keys           []*BlockchainTransactionKeyUpdate
}

type MempoolTransactionUpdate struct {
	Inserted               bool
	Tx                     *transaction.Transaction
	BlockchainNotification bool
	Keys                   map[string]bool
}
