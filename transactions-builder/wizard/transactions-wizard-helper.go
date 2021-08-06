package wizard

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/config_fees"
)

func setFeeTxNow(tx *transaction.Transaction, feePerByte uint64, value *uint64) {

	initAmount := *value

	var fee uint64
	oldFee := uint64(1)
	for oldFee != fee {
		oldFee = fee
		fee = config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())), feePerByte)
		*value = initAmount + fee
	}
	return
}

func setFeeFixedTxNow(fixedFee uint64, value *uint64) {
	*value = *value + fixedFee
}

func setFee(tx *transaction.Transaction, fixed, perByte uint64, perByteAuto bool, token []byte, payFeeInExtra bool) error {

	if fixed == 0 {

		if perByte == 0 && !perByteAuto {
			return nil
		}

		if perByte > 0 && perByteAuto {
			return errors.New("PerBye is set and PerByteAuto")
		}

		if perByte == 0 {
			if config_fees.FEES_PER_BYTE[string(token)] == 0 {
				return errors.New("The token will most like not be accepted by other miners")
			}
			perByte = config_fees.FEES_PER_BYTE[string(token)]
		}
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		if payFeeInExtra {

			switch base.TxScript {
			case transaction_simple.SCRIPT_UNSTAKE:
				if fixed > 0 {
					setFeeFixedTxNow(fixed, &base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
				} else {
					setFeeTxNow(tx, perByte, &base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
				}
				return nil
			}

		} else {

			for _, vin := range tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin {
				if bytes.Equal(vin.Token, token) {

					if fixed > 0 {
						setFeeFixedTxNow(fixed, &vin.Amount)
					} else {
						setFeeTxNow(tx, perByte, &vin.Amount)
					}
					return nil
				}
			}

			return errors.New("There is no input to set the fee!")
		}

	}

	return errors.New("Couldn't set fee")
}
