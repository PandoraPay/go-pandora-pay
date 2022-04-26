package wizard

import (
	"golang.org/x/exp/slices"
	"math"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/pending_stakes_list/pending_stakes"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

func CreateSimpleTx(transfer *WizardTxSimpleTransfer, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := transfer.Data.getData()
	if err != nil {
		return
	}

	spaceExtra := 0

	var txScript transaction_simple.ScriptType
	var extraFinal transaction_simple_extra.TransactionSimpleExtraInterface

	switch txExtra := transfer.Extra.(type) {
	case *WizardTxSimpleExtraUnstake:
		extraFinal = &transaction_simple_extra.TransactionSimpleExtraUnstake{
			Amounts: slices.Clone(txExtra.Amounts),
		}
		txScript = transaction_simple.SCRIPT_UNSTAKE
		spaceExtra += len(txExtra.Amounts) * len(helpers.SerializeToBytes(&pending_stakes.PendingStakes{nil, nil, math.MaxUint64, []*pending_stakes.PendingStake{{nil, helpers.RandomBytes(cryptography.PublicKeyHashSize), txExtra.Amounts[0], true}}}))
	}

	spaceExtra += len(transfer.Vout) * 50

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    txScript,
		DataVersion: transfer.Data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       transfer.Nonce,
		Extra:       extraFinal,
		Vin:         make([]*transaction_simple_parts.TransactionSimpleInput, len(transfer.Vin)),
		Vout:        make([]*transaction_simple_parts.TransactionSimpleOutput, len(transfer.Vout)),
	}

	privateKeys := make([]*addresses.PrivateKey, len(transfer.Vin))

	for i, vin := range transfer.Vin {
		if privateKeys[i], err = addresses.NewPrivateKey(transfer.Vin[i].Key); err != nil {
			return nil, err
		}
		txBase.Vin[i] = &transaction_simple_parts.TransactionSimpleInput{
			privateKeys[i].GeneratePublicKey(),
			vin.Amount,
			vin.Asset,
			nil,
		}
	}

	for i, vout := range transfer.Vout {
		txBase.Vout[i] = &transaction_simple_parts.TransactionSimpleOutput{
			vout.PublicKeyHash,
			vout.Amount,
			vout.Asset,
		}
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		SpaceExtra:               uint64(spaceExtra),
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	extraBytes := len(transfer.Vin) * cryptography.SignatureSize
	fee := setFee(tx, extraBytes, transfer.Fee.Clone(), true)
	if err = helpers.SafeUint64Add(&txBase.Vin[0].Amount, fee); err != nil {
		return nil, err
	}

	statusCallback("Transaction Fee set")

	statusCallback("Transaction Signing...")
	for i, vin := range txBase.Vin {
		if vin.Signature, err = privateKeys[i].Sign(tx.SerializeForSigning()); err != nil {
			return nil, err
		}
	}
	statusCallback("Transaction Signed")

	if err = bloomAllTx(tx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
