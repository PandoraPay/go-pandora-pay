package wizard

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/config_fees"
)

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

		if fee.PerByte == 0 && fee.PerByteAuto {
			fee.PerByte = config_fees.FEES_PER_BYTE
		}

	}

	if fee.Fixed > 0 {
		base.Fee = fee.Fixed
	} else if fee.PerByte > 0 {

		oldFee := uint64(0)
		for {
			fee := config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())), fee.PerByte)
			base.Fee = fee
			if oldFee == fee {
				break
			}
			oldFee = fee
		}

	}

	return nil
}

func setFeeZether(tx *transaction.Transaction, transfer *ZetherTransfer, fee *TransactionsWizardFee) error {

	if fee.Fixed == 0 {

		if fee.PerByte == 0 && !fee.PerByteAuto {
			return nil
		}
		if fee.PerByte > 0 && fee.PerByteAuto {
			return errors.New("PerBye is set and PerByteAuto")
		}

		if fee.PerByte == 0 {
			fee.PerByte = config_fees.FEES_PER_BYTE_ZETHER
		}

	}

	if fee.Fixed > 0 {
		transfer.Fee = fee.Fixed
	} else if fee.PerByte > 0 {
		fee := config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())), fee.PerByte)
		transfer.Fee = fee
	}

	return nil
}
