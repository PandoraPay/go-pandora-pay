package wizard

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/config_fees"
	"pandora-pay/helpers"
)

func setFee(tx *transaction.Transaction, extraBytes int, fee *TransactionsWizardFee, includeSerialize bool) uint64 {

	if fee.Fixed > 0 {
		return fee.Fixed
	}

	if fee.PerByte == 0 && fee.PerByteExtraSpace == 0 && !fee.PerByteAuto {
		return 0
	}

	if fee.PerByte == 0 && fee.PerByteAuto {
		fee.PerByte = config_fees.FEES_PER_BYTE_ZETHER
		fee.PerByteExtraSpace = config_fees.FEES_PER_BYTE_EXTRA_SPACE
	}

	spaceExtra := tx.SpaceExtra

	oldFee, feeValue := uint64(0), uint64(0)
	for {

		serializeLength := uint64(0)
		if includeSerialize {
			serializeLength = uint64(len(tx.SerializeManualToBytes()))
		}

		feeValue = config_fees.ComputeTxFees(serializeLength+uint64(helpers.BytesLengthSerialized(feeValue)+extraBytes), fee.PerByte, spaceExtra, fee.PerByteExtraSpace)

		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	return feeValue
}

func bloomAllTx(tx *transaction.Transaction, validateTx bool, statusCallback func(string)) (err error) {
	if validateTx {
		if err = tx.BloomAll(); err != nil {
			return
		}
		statusCallback("Transaction Bloomed")
		if err = tx.Verify(); err != nil {
			return
		}
		statusCallback("Transaction Verified")
	} else {
		if err = tx.BloomExtraVerified(); err != nil {
			return
		}
		if err = tx.BloomAll(); err != nil {
			return
		}
		statusCallback("Transaction Bloomed as Verified")
	}

	return
}
