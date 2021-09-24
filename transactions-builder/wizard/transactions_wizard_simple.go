package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
)

func signSimpleTransaction(tx *transaction.Transaction, privateKey *addresses.PrivateKey, statusCallback func(string)) (err error) {

	statusCallback("Transaction Signing...")

	if tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return err
	}

	statusCallback("Transaction Signed")

	return
}

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	privateKey := &addresses.PrivateKey{Key: key}

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	tx := &transaction.Transaction{
		Version:       transaction_type.TX_SIMPLE,
		Registrations: &transaction_data.TransactionDataTransactions{},
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			TxScript:    transaction_simple.SCRIPT_UNSTAKE,
			DataVersion: data.getDataVersion(),
			Data:        dataFinal,
			Nonce:       nonce,
			Fee:         0,
			TransactionSimpleExtraInterface: &transaction_simple_extra.TransactionSimpleUnstake{
				Amount: unstakeAmount,
			},
			Vin: &transaction_simple_parts.TransactionSimpleInput{
				PublicKey: privateKey.GeneratePublicKey(),
			},
		},
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return
	}
	if err = setFeeSimple(tx, fee.Clone()); err != nil {
		return
	}
	statusCallback("Transaction Fees set")

	if err = tx.BloomAll(); err != nil {
		return
	}
	statusCallback("Transaction Bloomed")

	if err = tx.Verify(); err != nil {
		return
	}
	statusCallback("Transaction Verified")

	return tx, nil
}

func CreateUpdateDelegateTx(nonce uint64, key []byte, delegateNewPubKey []byte, delegateNewFee uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	if len(delegateNewPubKey) != cryptography.PublicKeySize {
		return nil, errors.New("Delegating arguments are empty")
	}

	privateKey := &addresses.PrivateKey{Key: key}
	tx := &transaction.Transaction{
		Version:       transaction_type.TX_SIMPLE,
		Registrations: &transaction_data.TransactionDataTransactions{},
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			TxScript:    transaction_simple.SCRIPT_UPDATE_DELEGATE,
			DataVersion: data.getDataVersion(),
			Data:        dataFinal,
			Nonce:       nonce,
			TransactionSimpleExtraInterface: &transaction_simple_extra.TransactionSimpleUpdateDelegate{
				NewPublicKey: delegateNewPubKey,
				NewFee:       delegateNewFee,
			},
			Vin: &transaction_simple_parts.TransactionSimpleInput{
				PublicKey: privateKey.GeneratePublicKey(),
			},
		},
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return
	}

	if err = setFeeSimple(tx, fee.Clone()); err != nil {
		return
	}
	statusCallback("Transaction Fees set")

	if err = signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	statusCallback("Transaction Bloomed")

	if err = tx.Verify(); err != nil {
		return
	}
	statusCallback("Transaction Verified")

	return tx, nil
}

func CreateClaimTx(nonce uint64, key []byte, txRegistrations []*transaction_data.TransactionDataRegistration, output []*transaction_simple_parts.TransactionSimpleOutput, data *TransactionsWizardData, fee *TransactionsWizardFee, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := data.getData()
	if err != nil {
		return
	}

	privateKey := &addresses.PrivateKey{Key: key}
	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
		Registrations: &transaction_data.TransactionDataTransactions{
			Registrations: txRegistrations,
		},
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			TxScript:    transaction_simple.SCRIPT_UPDATE_DELEGATE,
			DataVersion: data.getDataVersion(),
			Data:        dataFinal,
			Nonce:       nonce,
			TransactionSimpleExtraInterface: &transaction_simple_extra.TransactionSimpleClaim{
				Output: output,
			},
			Vin: &transaction_simple_parts.TransactionSimpleInput{
				PublicKey: privateKey.GeneratePublicKey(),
			},
		},
	}
	statusCallback("Transaction Created")

	if err = signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return
	}

	if err = setFeeSimple(tx, fee.Clone()); err != nil {
		return
	}
	statusCallback("Transaction Fees set")

	if err = signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	statusCallback("Transaction Bloomed")

	if err = tx.Verify(); err != nil {
		return
	}
	statusCallback("Transaction Verified")

	return tx, nil
}
