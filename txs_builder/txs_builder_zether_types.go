package txs_builder

import (
	"pandora-pay/txs_builder/wizard"
)

type ZetherRingConfiguration struct {
	RingSize    int `json:"ringSize"  msgpack:"ringSize"`
	NewAccounts int `json:"newAccounts"  msgpack:"newAccounts"`
}

type TxBuilderCreateZetherTxPayload struct {
	Sender            string                             `json:"sender" msgpack:"sender"`
	Asset             []byte                             `json:"asset" msgpack:"asset"`
	Amount            uint64                             `json:"amount" msgpack:"amount"`
	Recipient         string                             `json:"recipient" msgpack:"recipient"`
	Burn              uint64                             `json:"burn" msgpack:"burn"`
	RingConfiguration *ZetherRingConfiguration           `json:"ringConfiguration" msgpack:"ringConfiguration"`
	Data              *wizard.WizardTransactionData      `json:"data" msgpack:"data"`
	Fee               *wizard.WizardZetherTransactionFee `json:"fee" msgpack:"fee"`
	Extra             wizard.WizardZetherPayloadExtra    `json:"extra" msgpack:"extra"`
}

type TxBuilderCreateZetherTxData struct {
	Payloads []*TxBuilderCreateZetherTxPayload `json:"payloads" msgpack:"payloads"`
}
