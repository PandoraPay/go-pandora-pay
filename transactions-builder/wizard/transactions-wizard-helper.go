package wizard

import (
	"encoding/binary"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_fees"
)

func setFee(tx *transaction.Transaction, fee *TransactionsWizardFee, setFeeCallback func(uint642 uint64)) error {

	if fee.Fixed > 0 {
		setFeeCallback(fee.Fixed)
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

	var n int
	buf := make([]byte, binary.MaxVarintLen64)

	oldFee, feeValue := uint64(0), uint64(0)
	for {

		n = binary.PutUvarint(buf, feeValue)

		feeValue = config_fees.ComputeTxFees(uint64(len(tx.SerializeManualToBytes())+n-1), fee.PerByte, tx.ComputeExtraSpace(), fee.PerByteExtraSpace)

		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	setFeeCallback(feeValue)
	return nil
}
