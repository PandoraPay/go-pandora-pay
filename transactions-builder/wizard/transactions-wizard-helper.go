package wizard

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/fees"
)

func setFeeTxNow(tx *transaction.Transaction, feePerByte, initAmount uint64, value *uint64) {
	var fee uint64
	oldFee := uint64(1)
	for oldFee != fee {
		oldFee = fee
		fee = fees.ComputeTxFees(uint64(len(tx.SerializeToBytes())), feePerByte)
		*value = initAmount + fee
	}
	return
}

func setFee(tx *transaction.Transaction, feePerByte int, feeToken []byte, payFeeInExtra bool) error {

	if feePerByte == 0 {
		return nil
	}

	if feePerByte == -1 {
		feePerByte = int(fees.FEES_PER_BYTE[string(feeToken)])
		if feePerByte == 0 {
			return errors.New("The token will most like not be accepted by other miners")
		}
	}

	if feePerByte != 0 {

		switch tx.TxType {
		case transaction_type.TxSimple:
			base := tx.TxBase.(*transaction_simple.TransactionSimple)

			if payFeeInExtra {

				switch base.TxScript {
				case transaction_simple.TxSimpleScriptUnstake:
					setFeeTxNow(tx, uint64(feePerByte), 0, &base.Extra.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
					return nil
				}

			} else {

				for _, vin := range tx.TxBase.(*transaction_simple.TransactionSimple).Vin {
					if bytes.Equal(vin.Token, feeToken) {
						setFeeTxNow(tx, uint64(feePerByte), vin.Amount, &vin.Amount)
						return nil
					}
				}

				return errors.New("There is no input to set the fee!")
			}

		}

	}

	return errors.New("Couldn't set fee")
}
