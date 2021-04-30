package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
)

func CreateSimpleTxOneInOneOut(nonce uint64, key []byte, amount uint64, token []byte, dst string, dstAmount uint64, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {
	return CreateSimpleTx(nonce, [][]byte{key}, []uint64{amount}, [][]byte{token}, []string{dst}, []uint64{dstAmount}, [][]byte{token}, feePerByte, feeToken)
}

func CreateSimpleTx(nonce uint64, keys [][]byte, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {

	if len(keys) != len(amounts) || len(amounts) != len(tokens) || len(amounts) == 0 {
		return nil, errors.New("Input lengths are a mismatch")
	}
	if len(dsts) != len(dstsAmounts) {
		return nil, errors.New("Output lengths are a mismatch")
	}

	privateKeys := make([]addresses.PrivateKey, len(keys))
	vin := make([]*transaction_simple.TransactionSimpleInput, len(keys))
	for i := 0; i < len(keys); i++ {

		privateKeys[i] = addresses.PrivateKey{Key: keys[i]}

		vin[i] = &transaction_simple.TransactionSimpleInput{
			Amount: amounts[i],
			Token:  tokens[i],
		}
	}

	vout := make([]*transaction_simple.TransactionSimpleOutput, len(dsts))
	for i := 0; i < len(dsts); i++ {
		var outAddress *addresses.Address
		if outAddress, err = addresses.DecodeAddr(dsts[i]); err != nil {
			return
		}
		vout[i] = &transaction_simple.TransactionSimpleOutput{
			PublicKeyHash: outAddress.PublicKeyHash,
			Amount:        dstsAmounts[i],
			Token:         dstsTokens[i],
		}
	}

	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TxSimple,
		TxBase: &transaction_simple.TransactionSimple{
			Nonce: nonce,
			Vin:   vin,
			Vout:  vout,
		},
	}

	for i, privateKey := range privateKeys {
		if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[i].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
			return
		}
	}
	if err = setFee(tx, feePerByte, feeToken, false); err != nil {
		return
	}
	for i, privateKey := range privateKeys {
		if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[i].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
			return
		}
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	if err = tx.Validate(); err != nil {
		return
	}
	if err = tx.Verify(); err != nil {
		return
	}
	return
}

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction, err error) {

	privateKey := addresses.PrivateKey{Key: key}
	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TxSimple,
		TxBase: &transaction_simple.TransactionSimple{
			TxScript: transaction_simple.TxSimpleScriptUnstake,
			Nonce:    nonce,
			Extra: &transaction_simple_extra.TransactionSimpleUnstake{
				Amount: unstakeAmount,
			},
			Vin: []*transaction_simple.TransactionSimpleInput{
				{
					Amount: 0,
				},
			},
		},
	}

	if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return
	}
	if err = setFee(tx, feePerByte, feeToken, payFeeInExtra); err != nil {
		return
	}
	if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	if err = tx.Validate(); err != nil {
		return
	}
	if err = tx.Verify(); err != nil {
		return
	}
	return
}

func CreateDelegateTx(nonce uint64, key []byte, delegateAmount uint64, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction, err error) {

	delegateHasNewPublicKeyHash := false
	var delegateNewPublicKeyHash []byte //33 byte
	if delegateNewPubKeyHash != nil {
		delegateHasNewPublicKeyHash = true
		delegateNewPublicKeyHash = delegateNewPubKeyHash
	}

	privateKey := addresses.PrivateKey{Key: key}
	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TxSimple,
		TxBase: &transaction_simple.TransactionSimple{
			TxScript: transaction_simple.TxSimpleScriptDelegate,
			Nonce:    nonce,
			Extra: &transaction_simple_extra.TransactionSimpleDelegate{
				Amount:              delegateAmount,
				HasNewPublicKeyHash: delegateHasNewPublicKeyHash,
				NewPublicKeyHash:    delegateNewPublicKeyHash,
			},
			Vin: []*transaction_simple.TransactionSimpleInput{
				{
					Amount: 0,
				},
			},
		},
	}

	if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return
	}
	if err = setFee(tx, feePerByte, feeToken, false); err != nil {
		return
	}
	if tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	if err = tx.Validate(); err != nil {
		return
	}
	if err = tx.Verify(); err != nil {
		return
	}
	return
}
