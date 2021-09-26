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

	if fee.Fixed > 0 {
		base.Fee = fee.Fixed
		return nil
	}

	if fee.PerByte == 0 && fee.PerByteExtraSpace == 0 && !fee.PerByteAuto {
		return nil
	}
	if (fee.PerByte > 0 || fee.PerByteExtraSpace > 0) && fee.PerByteAuto {
		return errors.New("PerBye is set and PerByteAuto")
	}

	if fee.PerByte == 0 && fee.PerByteAuto {
		fee.PerByte = config_fees.FEES_PER_BYTE
		fee.PerByteExtraSpace = config_fees.FEES_PER_BYTE_EXTRA_SPACE
	}

	oldFee := uint64(0)
	for {
		feeValue := config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())), fee.PerByte, uint64(64*len(tx.Registrations.Registrations)), fee.PerByteExtraSpace)
		base.Fee = feeValue
		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	return nil
}

func setFeeZether(tx *transaction.Transaction, transfer *ZetherTransfer, fee *TransactionsWizardFee, signCallback func() error) error {

	if fee.Fixed > 0 {
		transfer.Fee = fee.Fixed
		return nil
	}

	if fee.PerByte == 0 && fee.PerByteExtraSpace == 0 && !fee.PerByteAuto {
		return nil
	}
	if (fee.PerByte > 0 || fee.PerByteExtraSpace > 0) && fee.PerByteAuto {
		return errors.New("PerBye is set and PerByteAuto")
	}

	if fee.PerByte == 0 && fee.PerByteAuto {
		fee.PerByte = config_fees.FEES_PER_BYTE_ZETHER
		fee.PerByteExtraSpace = config_fees.FEES_PER_BYTE_EXTRA_SPACE
	}

	oldFee := uint64(0)
	for {
		feeValue := config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())), fee.PerByte, uint64(64*len(tx.Registrations.Registrations)), fee.PerByteExtraSpace)
		transfer.Fee = feeValue

		if err := signCallback(); err != nil {
			return err
		}

		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	return nil
}
