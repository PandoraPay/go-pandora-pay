package api_common

import "pandora-pay/blockchain/transactions/transaction"

type APIBlockchain struct {
	Height          uint64
	Hash            string
	PrevHash        string
	KernelHash      string
	PrevKernelHash  string
	Timestamp       uint64
	Transactions    uint64
	Target          string
	TotalDifficulty string
}

type APIBlockchainSync struct {
	SyncTime uint64
}

type APITransaction struct {
	Tx      *transaction.Transaction
	Mempool bool
}

type APITransactionSerialized struct {
	Tx      []byte
	Mempool bool
}
