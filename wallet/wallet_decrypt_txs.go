package wallet

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
)


type DecryptedTx struct {
	Type     transaction_type.TransactionVersion `json:"type" msgpack:"type"`
}

func (w *Wallet) DecryptTx(tx *transaction.Transaction, publicKeyHash []byte) (*DecryptedTx, error) {

	if tx == nil {
		return nil, errors.New("Transaction is invalid")
	}
	if err := tx.BloomAll(); err != nil {
		return nil, err
	}

	output := &DecryptedTx{
		Type: tx.Version,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
	}

	return output, nil
}
