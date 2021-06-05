package api_common

import (
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
)

type APIBlockWithTxs struct {
	Block *block.Block       `json:"block"`
	Txs   []helpers.HexBytes `json:"txs"`
}

type APIBlockchain struct {
	Height            uint64 `json:"height"`
	Hash              string `json:"hash"`
	PrevHash          string `json:"prevHash"`
	KernelHash        string `json:"kernelHash"`
	PrevKernelHash    string `json:"prevKernelHash"`
	Timestamp         uint64 `json:"timestamp"`
	TransactionsCount uint64 `json:"transactions"`
	Target            string `json:"target"`
	TotalDifficulty   string `json:"totalDifficulty"`
}

type APIBlockchainSync struct {
	SyncTime uint64 `json:"syncTime"`
}

type APITransaction struct {
	Tx      *transaction.Transaction `json:"tx,omitempty"`
	Mempool bool                     `json:"mempool,omitempty"`
}

type APITransactionSerialized struct {
	Tx      helpers.HexBytes `json:"tx,omitempty"`
	Mempool bool             `json:"mempool,omitempty"`
}

type APISubscriptionNotification struct {
	Key  helpers.HexBytes `json:"key,omitempty"`
	Data helpers.HexBytes `json:"tx,omitempty"`
}
