package wizard

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
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

func setFee(tx *transaction.Transaction, fee *TransactionsWizardFee) error {

	switch tx.Version {
	case transaction_type.TX_SIMPLE:

		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		if fee.Fixed == 0 {

			if fee.PerByte == 0 && !fee.PerByteAuto {
				return nil
			}

			if fee.PerByte > 0 && fee.PerByteAuto {
				return errors.New("PerBye is set and PerByteAuto")
			}

			if fee.PerByte == 0 {
				fee.PerByte = config_fees.FEES_PER_BYTE
			}
		}

		if fee.Fixed > 0 {
			setFeeFixedTxNow(fee.Fixed, &base.Fee)
		} else {
			setFeeTxNow(tx, fee.PerByte, &base.Fee)
		}

	}

	return errors.New("Couldn't set fee")
}
