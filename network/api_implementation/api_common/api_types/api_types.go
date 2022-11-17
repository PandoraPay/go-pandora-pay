package api_types

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
	Included bool `json:"included,omitempty" msgpack:"included,omitempty"`
}

type APISubscriptionNotificationAccountExtra struct {
	Asset []byte `json:"asset" msgpack:"asset"`
	Index uint64 `json:"index" msgpack:"index"`
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
	Included bool `json:"included,omitempty" msgpack:"included,omitempty"`
}

type APISubscriptionNotificationTxExtra struct {
	Blockchain *APISubscriptionNotificationTxExtraBlockchain `json:"blockchain,omitempty" msgpack:"blockchain,omitempty"`
	Mempool    *APISubscriptionNotificationTxExtraMempool    `json:"mempool,omitempty" msgpack:"mempool,omitempty"`
}
