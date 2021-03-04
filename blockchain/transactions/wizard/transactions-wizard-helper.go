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
		feePerByte = int(fees.FEES_PER_BYTE[string(feeToken)])
		if feePerByte == 0 {
			return errors.New("The token will most like not be accepted by other miners")
		}
	}

	if feePerByte != 0 {

		switch tx.TxType {
		case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:

			for _, vin := range tx.TxBase.(transaction_simple.TransactionSimple).Vin {
				if bytes.Equal(vin.Token, feeToken) {

					var initialAmount = vin.Amount

					var fee uint64
					oldFee := uint64(1)
					for oldFee != fee {
						fee = fees.ComputeTxFees(uint64(len(tx.Serialize(true))), uint64(feePerByte))
						oldFee = fee
						vin.Amount = initialAmount + fee
					}

					return
				}
			}

			return errors.New("There is no input to set the fee!")

		}

	}

	return
}
