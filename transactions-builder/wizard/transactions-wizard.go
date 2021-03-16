package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
)

func CreateSimpleTxOneInOneOut(nonce uint64, key []byte, amount uint64, token []byte, dst string, dstAmount uint64, feePerByte int, feeToken []byte) (tx *transaction.Transaction) {
	return CreateSimpleTx(nonce, [][]byte{key}, []uint64{amount}, [][]byte{token}, []string{dst}, []uint64{dstAmount}, [][]byte{token}, feePerByte, feeToken)
}

func CreateSimpleTx(nonce uint64, keys [][]byte, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction) {

	if len(keys) != len(amounts) || len(amounts) != len(tokens) || len(amounts) == 0 {
		panic("Input lengths are a mismatch")
	}
	if len(dsts) != len(dstsAmounts) {
		panic("Output lengths are a mismatch")
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
		outAddress := addresses.DecodeAddr(dsts[i])
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

	setFee(tx, feePerByte, feeToken, false)

	hash := tx.SerializeForSigning()
	for i, privateKey := range privateKeys {
		tx.TxBase.(*transaction_simple.TransactionSimple).Vin[i].Signature = privateKey.Sign(hash)
	}
	tx.BloomAll()
	tx.Validate()
	tx.Verify()
	return
}

func CreateUnstakeTx(nonce uint64, key []byte, unstakeAmount uint64, feePerByte int, feeToken []byte, payFeeInExtra bool) (tx *transaction.Transaction) {

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

	setFee(tx, feePerByte, feeToken, payFeeInExtra)
	tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature = privateKey.Sign(tx.SerializeForSigning())
	tx.BloomAll()
	tx.Validate()
	tx.Verify()
	return
}

func CreateDelegateTx(nonce uint64, key []byte, delegateAmount uint64, delegateNewPubKeyHash []byte, feePerByte int, feeToken []byte) (tx *transaction.Transaction) {

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

	setFee(tx, feePerByte, feeToken, false)
	tx.TxBase.(*transaction_simple.TransactionSimple).Vin[0].Signature = privateKey.Sign(tx.SerializeForSigning())
	tx.BloomAll()
	tx.Validate()
	tx.Verify()
	return
}
