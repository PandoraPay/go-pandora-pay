package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/cryptography"
)

func signSimpleTransaction(tx *transaction.Transaction, privateKey *addresses.PrivateKey, fee *TransactionsWizardFee, validateTx bool, statusCallback func(string)) (err error) {

	txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

	extraBytes := cryptography.SignatureSize
	txBase.Fees = setFee(tx, extraBytes, fee.Clone(), true)
	statusCallback("Transaction Fees set")

	statusCallback("Transaction Signing...")
	if txBase.Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return err
	}

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

	statusCallback("Transaction Signed")
	return
}

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	privateKey := &addresses.PrivateKey{Key: key}

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    transaction_simple.SCRIPT_UNSTAKE,
		DataVersion: data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       nonce,
		Fees:        0,
		Extra: &transaction_simple_extra.TransactionSimpleUnstake{
			Amount: unstakeAmount,
		},
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		Registrations:            &transaction_data.TransactionDataTransactions{},
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, fee, validateTx, statusCallback); err != nil {
		return
	}
	return tx, nil
}

func CreateUpdateDelegateTx(nonce uint64, key []byte, delegatedStakingNewPublicKey []byte, delegatedStakingNewFee, delegateStakingUpdateAmount uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	delegatedStakingHasNewInfo := false
	if len(delegatedStakingNewPublicKey) == cryptography.PublicKeySize {
		delegatedStakingHasNewInfo = true
	}

	if len(delegatedStakingNewPublicKey) != cryptography.PublicKeySize && delegatedStakingNewFee > 0 {
		return nil, errors.New("delegatedStakingNewFee is > 0 while the delegatedStakingNewPublicKey is not right")
	}

	privateKey := &addresses.PrivateKey{Key: key}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    transaction_simple.SCRIPT_UPDATE_DELEGATE,
		DataVersion: data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       nonce,
		Extra: &transaction_simple_extra.TransactionSimpleUpdateDelegate{
			DelegatedStakingUpdateAmount: delegateStakingUpdateAmount,
			DelegatedStakingHasNewInfo:   delegatedStakingHasNewInfo,
			DelegatedStakingNewPublicKey: delegatedStakingNewPublicKey,
			DelegatedStakingNewFee:       delegatedStakingNewFee,
		},
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		Registrations:            &transaction_data.TransactionDataTransactions{},
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, fee, validateTx, statusCallback); err != nil {
		return
	}
	return tx, nil
}

func CreateClaimTx(nonce uint64, key []byte, txRegistrations []*transaction_data.TransactionDataRegistration, output []*transaction_simple_parts.TransactionSimpleOutput, data *TransactionsWizardData, fee *TransactionsWizardFee, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	privateKey := &addresses.PrivateKey{Key: key}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    transaction_simple.SCRIPT_CLAIM,
		DataVersion: data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       nonce,
		Extra: &transaction_simple_extra.TransactionSimpleClaim{
			Output: output,
		},
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
	}

	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
		Registrations: &transaction_data.TransactionDataTransactions{
			Registrations: txRegistrations,
		},
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, fee, validateTx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
