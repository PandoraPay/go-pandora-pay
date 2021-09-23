package wizard

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/config_fees"
	"pandora-pay/helpers"
)

func setFeeTxSimpleNow(tx *transaction.Transaction, feePerByte uint64, value *uint64) (err error) {

	initAmount := *value

	var fee, oldFee uint64
	first := true
	for oldFee != fee || first {
		first = false

		oldFee = fee
		fee = config_fees.ComputeTxSimpleFees(uint64(len(tx.SerializeManualToBytes())), feePerByte)

		*value = initAmount
		if err = helpers.SafeUint64Add(value, fee); err != nil {
			return
		}
	}
	return
}

func setFeeSimple(tx *transaction.Transaction, fee *TransactionsWizardFee) error {

	if tx.Version != transaction_type.TX_SIMPLE {
		return errors.New("Tx Version is not TX SIMPLE")
	}

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
		return helpers.SafeUint64Add(&base.Fee, fee.Fixed)
	} else {
		return setFeeTxSimpleNow(tx, fee.PerByte, &base.Fee)
	}

}
