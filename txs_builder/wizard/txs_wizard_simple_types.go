package wizard

type WizardTxSimpleExtra interface {
}

type WizardTxSimpleExtraUnstake struct {
	WizardTxSimpleExtra
	Amounts []uint64
}

type WizardTxSimpleTransferVin struct {
	Key    []byte `json:"key" msgpack:"key"`
	Amount uint64 `json:"amount" msgpack:"amount"`
	Asset  []byte `json:"asset" msgpack:"asset"`
}

type WizardTxSimpleTransferVout struct {
	PublicKeyHash []byte `json:"publicKeyHash" msgpack:"publicKeyHash"`
	Amount        uint64 `json:"amount" msgpack:"amount"`
	Asset         []byte `json:"asset" msgpack:"asset"`
}

type WizardTxSimpleTransfer struct {
	Extra WizardTxSimpleExtra    `json:"extra" msgpack:"extra"`
	Data  *WizardTransactionData `json:"data" msgpack:"data"`
	Fee   *WizardTransactionFee  `json:"fee" msgpack:"fee"`

	Nonce uint64                        `json:"nonce" msgpack:"nonce"`
	Vin   []*WizardTxSimpleTransferVin  `json:"vin" msgpack:"vin"`
	Vout  []*WizardTxSimpleTransferVout `json:"vout" msgpack:"vout"`
}
