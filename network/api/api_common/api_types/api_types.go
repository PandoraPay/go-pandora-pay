package api_types

import (
	"pandora-pay/helpers"
)

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty" msgpack:"type,omitempty"`
	Key              helpers.HexBytes `json:"key,omitempty" msgpack:"key,omitempty"`
	Data             helpers.HexBytes `json:"data,omitempty" msgpack:"data,omitempty"`
	Extra            helpers.HexBytes `json:"extra,omitempty" msgpack:"extra,omitempty"`
}

type APISubscriptionNotificationTxExtraBlockchain struct {
	Inserted     bool   `json:"inserted,omitempty" msgpack:"inserted,omitempty"`
	BlkHeight    uint64 `json:"blkHeight" msgpack:"blkHeight"`
	BlkTimestamp uint64 `json:"blkTimestamp" msgpack:"blkTimestamp"`
	Height       uint64 `json:"height" msgpack:"height"`
}

type APISubscriptionNotificationAccountTxExtraBlockchain struct {
	Inserted     bool   `json:"inserted,omitempty" msgpack:"inserted,omitempty"`
	TxsCount     uint64 `json:"txsCount" msgpack:"txsCount"`
	BlkHeight    uint64 `json:"blkHeight" msgpack:"blkHeight"`
	BlkTimestamp uint64 `json:"blkTimestamp" msgpack:"blkTimestamp"`
	Height       uint64 `json:"height" msgpack:"height"`
}

type APISubscriptionNotificationAccountTxExtraMempool struct {
	Inserted bool `json:"inserted,omitempty" msgpack:"inserted,omitempty"`
}

type APISubscriptionNotificationAccountExtra struct {
	Asset helpers.HexBytes `json:"asset" msgpack:"asset"`
	Index uint64           `json:"index" msgpack:"index"`
}

type APISubscriptionNotificationPlainAccExtra struct {
	Index uint64 `json:"index" msgpack:"index"`
}

type APISubscriptionNotificationRegistrationExtra struct {
	Index uint64 `json:"index" msgpack:"index"`
}

type APISubscriptionNotificationAssetExtra struct {
	Index uint64 `json:"index" msgpack:"index"`
}

type APISubscriptionNotificationAccountTxExtra struct {
	Blockchain *APISubscriptionNotificationAccountTxExtraBlockchain `json:"blockchain,omitempty" msgpack:"blockchain,omitempty"`
	Mempool    *APISubscriptionNotificationAccountTxExtraMempool    `json:"mempool,omitempty" msgpack:"mempool,omitempty"`
}

type APISubscriptionNotificationTxExtraMempool struct {
	Inserted bool `json:"inserted,omitempty" msgpack:"inserted,omitempty"`
}

type APISubscriptionNotificationTxExtra struct {
	Blockchain *APISubscriptionNotificationTxExtraBlockchain `json:"blockchain,omitempty" msgpack:"blockchain,omitempty"`
	Mempool    *APISubscriptionNotificationTxExtraMempool    `json:"mempool,omitempty" msgpack:"mempool,omitempty"`
}
