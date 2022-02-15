package txs_builder

import "pandora-pay/txs_builder/wizard"

type TxBuilderCreateSimpleTx struct {
	Sender     string                        `json:"sender" msgpack:"sender"`
	Nonce      uint64                        `json:"nonce" msgpack:"nonce"`
	Data       *wizard.WizardTransactionData `json:"data" msgpack:"data"`
	Fee        *wizard.WizardTransactionFee  `json:"fee" msgpack:"fee"`
	FeeVersion bool                          `json:"feeVersion" msgpack:"feeVersion"`
	Extra      wizard.WizardTxSimpleExtra    `json:"extra" msgpack:"sender"`
}
