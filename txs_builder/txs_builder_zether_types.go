package txs_builder

import (
	"pandora-pay/txs_builder/txs_builder_zether_helper"
	"pandora-pay/txs_builder/wizard"
)

type ZetherSenderRingType struct {
	RequireStakedAccounts bool     `json:"requireStakedAccounts" msgpack:"requireStakedAccounts"`
	AvoidStakedAccounts   bool     `json:"avoidStakedAccounts" msgpack:"avoidStakedAccounts"`
	IncludeMembers        []string `json:"includeMembers" msgpack:"includeMembers"`
	NewAccounts           int      `json:"newAccounts" msgpack:"newAccounts"`
}

type ZetherRecipientRingType struct {
	RequireStakedAccounts bool     `json:"requireStakedAccounts" msgpack:"requireStakedAccounts"`
	AvoidStakedAccounts   bool     `json:"avoidStakedAccounts" msgpack:"avoidStakedAccounts"`
	IncludeMembers        []string `json:"includeMembers" msgpack:"includeMembers"`
	NewAccounts           int      `json:"newAccounts" msgpack:"newAccounts"`
}

type ZetherRingConfiguration struct {
	SenderRingType    *ZetherSenderRingType    `json:"senderRingType" msgpack:"senderRingType"`
	RecipientRingType *ZetherRecipientRingType `json:"recipientRingType" msgpack:"recipientRingType"`
}

type TxBuilderCreateZetherTxPayload struct {
	txs_builder_zether_helper.TxsBuilderZetherTxPayloadBase
	Asset             []byte                             `json:"asset" msgpack:"asset"`
	Amount            uint64                             `json:"amount" msgpack:"amount"`
	DecryptedBalance  uint64                             `json:"decryptedBalance" msgpack:"decryptedBalance"`
	RingConfiguration *ZetherRingConfiguration           `json:"ringConfiguration" msgpack:"ringConfiguration"`
	Burn              uint64                             `json:"burn" msgpack:"burn"`
	Data              *wizard.WizardTransactionData      `json:"data" msgpack:"data"`
	Fee               *wizard.WizardZetherTransactionFee `json:"fee" msgpack:"fee"`
	Extra             wizard.WizardZetherPayloadExtra    `json:"extra" msgpack:"extra"`
}

type TxBuilderCreateZetherTxData struct {
	Payloads []*TxBuilderCreateZetherTxPayload `json:"payloads" msgpack:"payloads"`
}
