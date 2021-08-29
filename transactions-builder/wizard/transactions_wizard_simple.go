package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
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

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, statusCallback func(string)) (*transaction.Transaction, error) {

	privateKey := &addresses.PrivateKey{Key: key}

	dataFinal, err := data.getData()
	if err != nil {
		return nil, err
	}

	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
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

	if err := signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return nil, err
	}

	if err := setFee(tx, fee); err != nil {
		return nil, err
	}
	statusCallback("Transaction Fees set")

	if err := signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return nil, err
	}

	if err := tx.BloomAll(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Bloomed")

	if err := tx.Validate(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Validated")

	if err := tx.Verify(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Verified")

	return tx, nil
}

func CreateUpdateDelegateTx(nonce uint64, key []byte, delegateNewPubKey []byte, delegateNewFee uint64, data *TransactionsWizardData, fee *TransactionsWizardFee, statusCallback func(string)) (*transaction.Transaction, error) {

	dataFinal, err := data.getData()
	if err != nil {
		return nil, err
	}

	if len(delegateNewPubKey) != cryptography.PublicKeySize {
		return nil, errors.New("Delegating arguments are empty")
	}

	privateKey := &addresses.PrivateKey{Key: key}
	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
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

	if err := signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return nil, err
	}

	if err := setFee(tx, fee); err != nil {
		return nil, err
	}
	statusCallback("Transaction Fees set")

	if err := signSimpleTransaction(tx, privateKey, statusCallback); err != nil {
		return nil, err
	}

	if err := tx.BloomAll(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Bloomed")

	if err := tx.Validate(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Validated")

	if err := tx.Verify(); err != nil {
		return nil, err
	}
	statusCallback("Transaction Verified")

	return tx, nil
}
