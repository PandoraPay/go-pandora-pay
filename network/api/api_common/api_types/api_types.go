package api_types

import (
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
)

func GetReturnType(s string, defaultValue APIReturnType) APIReturnType {
	switch s {
	case "0":
		return APIReturnType_RETURN_JSON
	case "1":
		return APIReturnType_RETURN_SERIALIZED
	default:
		return defaultValue
	}
}

type APIBlockWithTxs struct {
	Block           *block.Block       `json:"block,omitempty"`
	BlockSerialized helpers.HexBytes   `json:"serialized,omitempty"`
	Txs             []helpers.HexBytes `json:"txs,omitempty"`
}

type APIAccount struct {
	Assets             []helpers.HexBytes          `json:"assets,omitempty"`
	Accs               []*account.Account          `json:"accounts,omitempty"`
	AccsSerialized     []helpers.HexBytes          `json:"accountsSerialized,omitempty"`
	PlainAcc           *plain_account.PlainAccount `json:"plainAccount,omitempty"`
	PlainAccSerialized helpers.HexBytes            `json:"plainAccountSerialized,omitempty"`
	Reg                *registration.Registration  `json:"registration,omitempty"`
	RegSerialized      helpers.HexBytes            `json:"registrationSerialized,omitempty"`
}

type APIAccountsByKeys struct {
	Acc           []*account.Account           `json:"acc,omitempty"`
	AccSerialized []helpers.HexBytes           `json:"accSerialized,omitempty"`
	Reg           []*registration.Registration `json:"registration,omitempty"`
	RegSerialized []helpers.HexBytes           `json:"registrationSerialized,omitempty"`
}

type APITransaction struct {
	Tx           *transaction.Transaction `json:"tx,omitempty"`
	TxSerialized helpers.HexBytes         `json:"serialized,omitempty"`
	Mempool      bool                     `json:"mempool,omitempty"`
	Info         *info.TxInfo             `json:"info,omitempty"`
}

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
