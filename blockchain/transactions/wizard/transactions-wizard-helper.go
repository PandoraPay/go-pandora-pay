package wizard

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/fees"
)

func setFee(tx *transaction.Transaction, feePerByte int, feeToken []byte) (err error) {

	if feePerByte == -1 {
		feePerByte = fees.FEE_PER_BYTE_DEFAULT
	}

	if feePerByte != 0 {

		switch tx.TxType {
		case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:

			var vinFee *transaction_simple.TransactionSimpleInput
			for _, vin := range tx.TxBase.(transaction_simple.TransactionSimple).Vin {
				if bytes.Equal(vin.Token, feeToken) {
					vinFee = &vin
					break
				}
			}

			if vinFee == nil {
				err = errors.New("There is no input to set the fee!")
				return
			}

			var initialAmount = vinFee.Amount

			var fee uint64
			oldFee := uint64(1)
			for oldFee != fee {
				fee = fees.ComputeTxFees(uint64(len(tx.Serialize(true))), uint64(feePerByte))
				oldFee = fee
				vinFee.Amount = initialAmount + fee
			}

		}

	}

	return
}
