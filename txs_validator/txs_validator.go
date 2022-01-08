package txs_validator

type TxsValidator struct {
}

func NewTxsValidator() (*TxsValidator, error) {
	txsValidator := &TxsValidator{}
	return txsValidator, nil
}
