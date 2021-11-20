package api_types

import (
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
)

type APITransactionPreview struct {
	TxPreview *info.TxPreview `json:"txPreview,omitempty"`
	Mempool   bool            `json:"mempool,omitempty"`
	Info      *info.TxInfo    `json:"info,omitempty"`
}

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty"`
	Key              helpers.HexBytes `json:"key,omitempty"`
	Data             helpers.HexBytes `json:"data,omitempty"`
	Extra            helpers.HexBytes `json:"extra,omitempty"`
}

type APISubscriptionNotificationTxExtraBlockchain struct {
	Inserted     bool   `json:"inserted,omitempty"`
	BlkHeight    uint64 `json:"blkHeight,omitempty"`
	BlkTimestamp uint64 `json:"blkTimestamp,omitempty"`
	Height       uint64 `json:"height,omitempty"`
}

type APISubscriptionNotificationAccountTxExtraBlockchain struct {
	Inserted     bool   `json:"inserted,omitempty"`
	TxsCount     uint64 `json:"txsCount,omitempty"`
	BlkHeight    uint64 `json:"blkHeight,omitempty"`
	BlkTimestamp uint64 `json:"blkTimestamp,omitempty"`
	Height       uint64 `json:"height,omitempty"`
}

type APISubscriptionNotificationAccountTxExtraMempool struct {
	Inserted bool `json:"inserted,omitempty"`
}

type APISubscriptionNotificationAccountExtra struct {
	Asset helpers.HexBytes `json:"asset"`
	Index uint64           `json:"index"`
}

type APISubscriptionNotificationPlainAccExtra struct {
	Index uint64 `json:"index"`
}

type APISubscriptionNotificationRegistrationExtra struct {
	Index uint64 `json:"index"`
}

type APISubscriptionNotificationAssetExtra struct {
	Index uint64 `json:"index"`
}

type APISubscriptionNotificationAccountTxExtra struct {
	Blockchain *APISubscriptionNotificationAccountTxExtraBlockchain `json:"blockchain,omitempty"`
	Mempool    *APISubscriptionNotificationAccountTxExtraMempool    `json:"mempool,omitempty"`
}

type APISubscriptionNotificationTxExtraMempool struct {
	Inserted bool `json:"inserted,omitempty"`
}

type APISubscriptionNotificationTxExtra struct {
	Blockchain *APISubscriptionNotificationTxExtraBlockchain `json:"blockchain,omitempty"`
	Mempool    *APISubscriptionNotificationTxExtraMempool    `json:"mempool,omitempty"`
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
