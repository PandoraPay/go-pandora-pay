package api_types

import (
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
)

type APIBlockCompleteMissingTxs struct {
	Txs []helpers.HexBytes `json:"txs,omitempty"`
}

type APIBlockWithTxs struct {
	Block           *block.Block       `json:"block,omitempty"`
	BlockSerialized helpers.HexBytes   `json:"serialized,omitempty"`
	Txs             []helpers.HexBytes `json:"txs,omitempty"`
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
	Tx           *transaction.Transaction `json:"tx,omitempty"`
	TxSerialized helpers.HexBytes         `json:"serialized,omitempty"`
	Mempool      bool                     `json:"mempool,omitempty"`
	Info         *info.TxInfo             `json:"info,omitempty"`
}

type APIMempoolAnswer struct {
	Count  int                `json:"count"`
	Hashes []helpers.HexBytes `json:"hashes"`
}

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty"`
	Key              helpers.HexBytes `json:"key,omitempty"`
	Data             helpers.HexBytes `json:"data,omitempty"`
	Extra            helpers.HexBytes `json:"extra,omitempty"`
}

type APISubscriptionNotificationAccountTxExtraBlockchain struct {
	Inserted bool   `json:"inserted,omitempty"`
	TxsCount uint64 `json:"txsCount,omitempty"`
}

type APISubscriptionNotificationAccountTxExtraMempool struct {
	Inserted bool   `json:"inserted,omitempty"`
}

type APISubscriptionNotificationAccountTxExtra struct {
	Blockchain *APISubscriptionNotificationAccountTxExtraBlockchain `json:"blockchain,omitempty"`
	Mempool *APISubscriptionNotificationAccountTxExtraMempool `json:"mempool,omitempty"`
}


type APISubscriptionNotificationTxExtraBlockchain struct {
	Inserted     bool   `json:"inserted,omitempty"`
	BlkHeight    uint64 `json:"blkHeight,omitempty"`
	BlkTimestamp uint64 `json:"blkTimestamp,omitempty"`
	Height       uint64 `json:"height,omitempty"`
}

type APISubscriptionNotificationTxExtraMempool struct {
	Inserted bool   `json:"inserted,omitempty"`
}

type APISubscriptionNotificationTxExtra struct {
	Blockchain *APISubscriptionNotificationTxExtraBlockchain `json:"blockchain,omitempty"`
	Mempool *APISubscriptionNotificationTxExtraMempool `json:"mempool,omitempty"`
}

type APIAccountTxs struct {
	Count uint64             `json:"count,omitempty"`
	Txs   []helpers.HexBytes `json:"txs,omitempty"`
}

type APIFaucetInfo struct {
	HCaptchaSiteKey      string `json:"hCaptchaSiteKey,omitempty"`
	FaucetTestnetEnabled bool   `json:"faucetTestnetEnabled,omitempty"`
	FaucetTestnetCoins   uint64 `json:"faucetTestnetCoins,omitempty"`
}
