package info

import (
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
)

type TxPreviewZetherPayloadExtraStakingReward struct {
	Reward uint64 `json:"reward" msgpack:"reward"`
}

type TxPreviewZetherPayloadExtraStaking struct {
}

type TxPreviewZetherPayloadExtraSpend struct {
}

type TxPreviewZetherPayloadExtraPayToScript struct {
	Deadline          uint64 `json:"deadline" msgpack:"dealine"`
	DefaultResolution bool   `json:"defaultResolution" msgpack:"defaultResolution"`
	Threshold         byte   `json:"threshold" msgpack:"threshold"`
}

type TxPreviewZetherPayload struct {
	PayloadScript transaction_zether_payload_script.PayloadScriptType `json:"payloadScript" msgpack:"payloadScript"`
	Asset         []byte                                              `json:"asset" msgpack:"asset"`
	BurnValue     uint64                                              `json:"burnValue" msgpack:"burnValue"`
	DataVersion   transaction_data.TransactionDataVersion             `json:"dataVersion" msgpack:"dataVersion"`
	DataPublic    []byte                                              `json:"dataPublic" msgpack:"dataPublic"`
	Ring          byte                                                `json:"ring" msgpack:"ring"`
	Extra         any                                                 `json:"extra" msgpack:"extra"`
}

type TxPreviewZether struct {
	Payloads []*TxPreviewZetherPayload `json:"payloads"  msgpack:"payloads"`
}
