package wizard

type WizardTxSimpleExtra interface {
}

type WizardTxSimpleTransfer struct {
	Extra  WizardTxSimpleExtra    `json:"extra" msgpack:"extra"`
	Data   *WizardTransactionData `json:"data" msgpack:"data"`
	Fee    *WizardTransactionFee  `json:"fee" msgpack:"fee"`
	Nonce  uint64                 `json:"nonce" msgpack:"nonce"`
	VinKey []byte                 `json:"vinKey" msgpack:"vinKey"`
}
