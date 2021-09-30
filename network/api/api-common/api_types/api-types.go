package api_types

import (
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data_storage/accounts/account"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
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

type APIAccount struct {
	Tokens             []helpers.HexBytes          `json:"tokens,omitempty"`
	Accs               []*account.Account          `json:"accounts,omitempty"`
	AccsSerialized     []helpers.HexBytes          `json:"accountsSerialized,omitempty"`
	PlainAcc           *plain_account.PlainAccount `json:"plainAccount,omitempty"`
	PlainAccSerialized helpers.HexBytes            `json:"plainAccountSerialized,omitempty"`
	Reg                *registration.Registration  `json:"registration,omitempty"`
	RegSerialized      helpers.HexBytes            `json:"registrationSerialized,omitempty"`
}

type APITransaction struct {
	Tx           *transaction.Transaction `json:"tx,omitempty"`
	TxSerialized helpers.HexBytes         `json:"serialized,omitempty"`
	Mempool      bool                     `json:"mempool,omitempty"`
	Info         *info.TxInfo             `json:"info,omitempty"`
}

type APIMempoolAnswer struct {
	ChainHash helpers.HexBytes   `json:"chainHash"`
	Count     int                `json:"count"`
	Hashes    []helpers.HexBytes `json:"hashes"`
}

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty"`
	Key              helpers.HexBytes `json:"key,omitempty"`
	Data             helpers.HexBytes `json:"data_storage,omitempty"`
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
	Token helpers.HexBytes `json:"token"`
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
