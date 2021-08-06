package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
)

func signSimpleTransaction(tx *transaction.Transaction, privateKeys []*addresses.PrivateKey) (err error) {

	for i, privateKey := range privateKeys {
		if tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Vin[i].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
			return err
		}
	}

	return
}

func CreateSimpleTxOneInOneOut(nonce uint64, key []byte, amount uint64, token []byte, dst string, dstAmount uint64, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, statusCallback func(string)) (*transaction.Transaction, error) {
	return CreateSimpleTx(nonce, [][]byte{key}, []uint64{amount}, [][]byte{token}, []string{dst}, []uint64{dstAmount}, [][]byte{token}, feeFixed, feePerByte, feePerByteAuto, feeToken, statusCallback)
}

func CreateSimpleTx(nonce uint64, keys [][]byte, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, statusCallback func(string)) (*transaction.Transaction, error) {

	if len(keys) != len(amounts) || len(amounts) != len(tokens) || len(amounts) == 0 {
		return nil, errors.New("Input lengths are a mismatch")
	}
	if len(dsts) != len(dstsAmounts) {
		return nil, errors.New("Output lengths are a mismatch")
	}

	privateKeys := make([]*addresses.PrivateKey, len(keys))
	vin := make([]*transaction_simple_parts.TransactionSimpleInput, len(keys))
	for i := 0; i < len(keys); i++ {

		privateKeys[i] = &addresses.PrivateKey{Key: keys[i]}

		vin[i] = &transaction_simple_parts.TransactionSimpleInput{
			Amount: amounts[i],
			Token:  tokens[i],
		}
	}

	vout := make([]*transaction_simple_parts.TransactionSimpleOutput, len(dsts))
	for i := 0; i < len(dsts); i++ {
		outAddress, err := addresses.DecodeAddr(dsts[i])
		if err != nil {
			return nil, err
		}
		vout[i] = &transaction_simple_parts.TransactionSimpleOutput{
			PublicKeyHash: outAddress.PublicKeyHash,
			Amount:        dstsAmounts[i],
			Token:         dstsTokens[i],
		}
	}

	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			Nonce: nonce,
			Vin:   vin,
			Vout:  vout,
		},
	}

	statusCallback("Transaction created")

	if err := signSimpleTransaction(tx, privateKeys); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #1")

	if err := setFee(tx, feeFixed, feePerByte, feePerByteAuto, feeToken, false); err != nil {
		return nil, err
	}
	statusCallback("Transaction Fees set")

	if err := signSimpleTransaction(tx, privateKeys); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #2")

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

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, feePayInExtra bool, statusCallback func(string)) (*transaction.Transaction, error) {

	privateKey := &addresses.PrivateKey{Key: key}

	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			TxScript: transaction_simple.SCRIPT_UNSTAKE,
			Nonce:    nonce,
			TransactionSimpleExtraInterface: &transaction_simple_extra.TransactionSimpleUnstake{
				Amount: unstakeAmount,
			},
			Vin: []*transaction_simple_parts.TransactionSimpleInput{
				{
					Amount: 0,
				},
			},
		},
	}
	statusCallback("Transaction Created")

	if err := signSimpleTransaction(tx, []*addresses.PrivateKey{privateKey}); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #1")

	if err := setFee(tx, feeFixed, feePerByte, feePerByteAuto, feeToken, feePayInExtra); err != nil {
		return nil, err
	}
	statusCallback("Transaction Fees set")

	if err := signSimpleTransaction(tx, []*addresses.PrivateKey{privateKey}); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #2")

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

func CreateDelegateTx(nonce uint64, key []byte, delegateAmount uint64, delegateNewPubKeyHash []byte, feeFixed, feePerByte uint64, feePerByteAuto bool, feeToken []byte, statusCallback func(string)) (*transaction.Transaction, error) {

	delegateHasNewPublicKeyHash := false
	var delegateNewPublicKeyHash []byte //33 byte
	if delegateNewPubKeyHash != nil {
		delegateHasNewPublicKeyHash = true
		delegateNewPublicKeyHash = delegateNewPubKeyHash
	}

	privateKey := &addresses.PrivateKey{Key: key}
	tx := &transaction.Transaction{
		Version: transaction_type.TX_SIMPLE,
		TransactionBaseInterface: &transaction_simple.TransactionSimple{
			TxScript: transaction_simple.SCRIPT_DELEGATE,
			Nonce:    nonce,
			TransactionSimpleExtraInterface: &transaction_simple_extra.TransactionSimpleDelegate{
				Amount:              delegateAmount,
				HasNewPublicKeyHash: delegateHasNewPublicKeyHash,
				NewPublicKeyHash:    delegateNewPublicKeyHash,
			},
			Vin: []*transaction_simple_parts.TransactionSimpleInput{
				{
					Amount: 0,
				},
			},
		},
	}
	statusCallback("Transaction Created")

	if err := signSimpleTransaction(tx, []*addresses.PrivateKey{privateKey}); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #1")

	if err := setFee(tx, feeFixed, feePerByte, feePerByteAuto, feeToken, false); err != nil {
		return nil, err
	}
	statusCallback("Transaction Fees set")

	if err := signSimpleTransaction(tx, []*addresses.PrivateKey{privateKey}); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed #2")

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
