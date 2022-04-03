package txs_builder

import "pandora-pay/txs_builder/wizard"

type TxBuilderCreateSimpleTxVin struct {
	Sender string `json:"sender" msgpack:"sender"`
	Amount uint64 `json:"amount" msgpack:"amount"`
	Asset  []byte `json:"asset" msgpack:"asset"`
}

type TxBuilderCreateSimpleTxVout struct {
	PublicKeyHash []byte `json:"publicKeyHash" msgpack:"publicKeyHash"`
	Amount        uint64 `json:"amount" msgpack:"amount"`
	Asset         []byte `json:"asset" msgpack:"asset"`
}

type TxBuilderCreateSimpleTx struct {
	Nonce      uint64                         `json:"nonce" msgpack:"nonce"`
	Data       *wizard.WizardTransactionData  `json:"data" msgpack:"data"`
	Fee        *wizard.WizardTransactionFee   `json:"fee" msgpack:"fee"`
	FeeVersion bool                           `json:"feeVersion" msgpack:"feeVersion"`
	Extra      wizard.WizardTxSimpleExtra     `json:"extra" msgpack:"sender"`
	Vin        []*TxBuilderCreateSimpleTxVin  `json:"vin" msgpack:"vin"`
	Vout       []*TxBuilderCreateSimpleTxVout `json:"vout" msgpack:"vout"`
}
