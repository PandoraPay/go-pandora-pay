package wizard

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
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

func setFee(tx *transaction.Transaction, fee *TransactionsWizardFeeExtra) error {

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
				if config_fees.FEES_PER_BYTE[string(base.Token)] == 0 {
					return errors.New("The token will most like not be accepted by other miners")
				}
				fee.PerByte = config_fees.FEES_PER_BYTE[string(base.Token)]
			}
		}

		if fee.PayInExtra {

			if !bytes.Equal(base.Token, config.NATIVE_TOKEN) {
				return errors.New("Pay In Extra can not be paid in a different token")
			}

			switch base.TxScript {
			case transaction_simple.SCRIPT_UNSTAKE:
				if fee.Fixed > 0 {
					setFeeFixedTxNow(fee.Fixed, &base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
				} else {
					setFeeTxNow(tx, fee.PerByte, &base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake).FeeExtra)
				}
				return nil
			}

		} else {

			for _, vin := range base.Vin {
				if fee.Fixed > 0 {
					setFeeFixedTxNow(fee.Fixed, &vin.Amount)
				} else {
					setFeeTxNow(tx, fee.PerByte, &vin.Amount)
				}
				return nil
			}

			return errors.New("There is no input to set the fee!")
		}

	}

	return errors.New("Couldn't set fee")
}
