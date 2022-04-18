package wizard

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/config/config_fees"
	"pandora-pay/helpers"
)

func setFee(tx *transaction.Transaction, extraBytes int, fee *WizardTransactionFee, includeSerialize bool) uint64 {

	if fee.Fixed > 0 {
		return fee.Fixed
	}

	if fee.PerByte == 0 && fee.PerByteExtraSpace == 0 && !fee.PerByteAuto {
		return 0
	}

	if fee.PerByte == 0 && fee.PerByteAuto {
		switch tx.Version {
		case transaction_type.TX_SIMPLE:
			fee.PerByte = config_fees.FEE_PER_BYTE
		}
		fee.PerByteExtraSpace = config_fees.FEE_PER_BYTE_EXTRA_SPACE
	}

	spaceExtra := tx.SpaceExtra

	oldFee, feeValue := uint64(0), uint64(0)
	for {

		serializeLength := uint64(0)
		if includeSerialize {
			serializeLength = uint64(len(tx.SerializeManualToBytes()))
		}

		feeValue = config_fees.ComputeTxFee(serializeLength+uint64(helpers.BytesLengthSerialized(feeValue)+extraBytes), fee.PerByte, spaceExtra, fee.PerByteExtraSpace)

		if oldFee == feeValue {
			break
		}
		oldFee = feeValue
	}

	return feeValue
}

func bloomAllTx(tx *transaction.Transaction, statusCallback func(string)) (err error) {

	if err = tx.BloomAll(); err != nil {
		return
	}
	statusCallback("Transaction Bloomed")
	if err = tx.Verify(); err != nil {
		return
	}
	statusCallback("Transaction Verified")

	return
}
