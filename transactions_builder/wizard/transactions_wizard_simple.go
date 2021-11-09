package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

func signSimpleTransaction(tx *transaction.Transaction, privateKey *addresses.PrivateKey, fee *TransactionsWizardFee, statusCallback func(string)) (err error) {

	txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

	extraBytes := cryptography.SignatureSize
	txBase.Fee = setFee(tx, extraBytes, fee.Clone(), true)
	statusCallback("Transaction Fee set")

	statusCallback("Transaction Signing...")
	if txBase.Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return err
	}
	statusCallback("Transaction Signed")

	return
}

func CreateSimpleTx(nonce uint64, key []byte, chainHeight uint64, extra WizardTxSimpleExtra, data *TransactionsWizardData, fee *TransactionsWizardFee, feeVersion bool, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	privateKey := &addresses.PrivateKey{Key: key}

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	spaceExtra := 0

	var txScript transaction_simple.ScriptType
	var extraFinal transaction_simple_extra.TransactionSimpleExtraInterface
	switch txExtra := extra.(type) {
	case *WizardTxSimpleExtraUpdateDelegate:
		extraFinal = &transaction_simple_extra.TransactionSimpleExtraUpdateDelegate{
			DelegatedStakingClaimAmount: txExtra.DelegatedStakingClaimAmount,
			DelegatedStakingUpdate:      txExtra.DelegatedStakingUpdate,
		}
		txScript = transaction_simple.SCRIPT_UPDATE_DELEGATE

		if txExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo {
			spaceExtra += len(txExtra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey)
			spaceExtra += helpers.BytesLengthSerialized(txExtra.DelegatedStakingUpdate.DelegatedStakingNewFee)
		}
		if txExtra.DelegatedStakingClaimAmount > 0 {
			spaceExtra += len(helpers.SerializeToBytes(&dpos.DelegatedStakePending{nil, txExtra.DelegatedStakingClaimAmount, chainHeight + 100, dpos.DelegatedStakePendingStake}))
		}
	case *WizardTxSimpleExtraUnstake:
		extraFinal = &transaction_simple_extra.TransactionSimpleExtraUnstake{
			Amount: txExtra.Amount,
		}
		txScript = transaction_simple.SCRIPT_UNSTAKE

		spaceExtra += len(helpers.SerializeToBytes(&dpos.DelegatedStakePending{nil, txExtra.Amount, chainHeight + 100, dpos.DelegatedStakePendingUnstake}))
	}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    txScript,
		DataVersion: data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       nonce,
		Fee:         0,
		FeeVersion:  feeVersion,
		Extra:       extraFinal,
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		SpaceExtra:               uint64(spaceExtra),
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, fee, statusCallback); err != nil {
		return
	}
	if err = bloomAllTx(tx, validateTx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
