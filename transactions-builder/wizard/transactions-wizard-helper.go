package wizard

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_fees"
)

func setFee(tx *transaction.Transaction, fee *TransactionsWizardFee, setFeeCallback func(fee uint64), signCallback func() error) error {

	if fee.Fixed > 0 {
		setFeeCallback(fee.Fixed)
		return signCallback()
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

		setFeeCallback(feeValue)

		if err := signCallback(); err != nil {
			return err
		}

		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	return signCallback()
}
