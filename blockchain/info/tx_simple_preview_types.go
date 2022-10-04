package info

import (
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
)

type TxPreviewSimpleExtraResolutionConditionalPayment struct {
	TxId         []byte `json:"txId" msgpack:"txId"`
	PayloadIndex byte   `json:"payloadIndex" msgpack:"payloadIndex"`
	Resolution   bool   `json:"resolution" msgpack:"resolution"`
}

type TxPreviewSimple struct {
	TxScript    transaction_simple.ScriptType           `json:"txScript" msgpack:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion" msgpack:"dataVersion"`
	DataPublic  []byte                                  `json:"dataPublic" msgpack:"dataPublic"`
	Vin         []byte                                  `json:"vin" msgpack:"vin"`
	Extra       any                                     `json:"extra" msgpack:"extra"`
}
