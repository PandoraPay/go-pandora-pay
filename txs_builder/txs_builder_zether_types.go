package txs_builder

import (
	"pandora-pay/txs_builder/wizard"
)

type ZetherSenderRingType struct {
	CopyRingMembers       int      `json:"copyRingMembers" msgpack:"copyRingMembers"`
	RequireStakedAccounts bool     `json:"requireStakedAccounts" msgpack:"requireStakedAccounts"`
	AvoidStakedAccounts   bool     `json:"avoidStakedAccounts" msgpack:"avoidStakedAccounts"`
	IncludeMembers        []string `json:"includeMembers" msgpack:"includeMembers"`
	NewAccounts           int      `json:"newAccounts" msgpack:"newAccounts"`
}

type ZetherRecipientRingType struct {
	CopyRingMembers       int      `json:"copyRingMembers" msgpack:"copyRingMembers"`
	RequireStakedAccounts bool     `json:"requireStakedAccounts" msgpack:"requireStakedAccounts"`
	AvoidStakedAccounts   bool     `json:"avoidStakedAccounts" msgpack:"avoidStakedAccounts"`
	IncludeMembers        []string `json:"includeMembers" msgpack:"includeMembers"`
	NewAccounts           int      `json:"newAccounts" msgpack:"newAccounts"`
}

type ZetherRingConfiguration struct {
	RingSize          int                      `json:"ringSize"  msgpack:"ringSize"`
	SenderRingType    *ZetherSenderRingType    `json:"senderRingType" msgpack:"senderRingType"`
	RecipientRingType *ZetherRecipientRingType `json:"recipientRingType" msgpack:"recipientRingType"`
}

type TxBuilderCreateZetherTxPayload struct {
	Sender            string                             `json:"sender" msgpack:"sender"`
	Asset             []byte                             `json:"asset" msgpack:"asset"`
	Amount            uint64                             `json:"amount" msgpack:"amount"`
	DecryptedBalance  uint64                             `json:"decryptedBalance" msgpack:"decryptedBalance"`
	Recipient         string                             `json:"recipient" msgpack:"recipient"`
	Burn              uint64                             `json:"burn" msgpack:"burn"`
	RingConfiguration *ZetherRingConfiguration           `json:"ringConfiguration" msgpack:"ringConfiguration"`
	Data              *wizard.WizardTransactionData      `json:"data" msgpack:"data"`
	Fee               *wizard.WizardZetherTransactionFee `json:"fee" msgpack:"fee"`
	Extra             wizard.WizardZetherPayloadExtra    `json:"extra" msgpack:"extra"`
	WitnessIndexes    []int                              `json:"witnessIndexes" msgpack:"witnessIndexes"`
}

type TxBuilderCreateZetherTxData struct {
	Payloads []*TxBuilderCreateZetherTxPayload `json:"payloads" msgpack:"payloads"`
}
