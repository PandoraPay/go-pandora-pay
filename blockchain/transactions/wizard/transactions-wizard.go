package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction_simple_unstake"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/fees"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
)

func CreateSimpleTxOneInOneOut(nonce uint64, key [32]byte, amount uint64, token []byte, dst string, dstAmount uint64, fee, computeFeeBlockHeight uint64) (tx *transaction.Transaction, err error) {
	return CreateSimpleTx(nonce, [][32]byte{key}, []uint64{amount}, [][]byte{token}, []string{dst}, []uint64{dstAmount}, [][]byte{token}, fee, computeFeeBlockHeight)
}

func CreateSimpleTx(nonce uint64, keys [][32]byte, amounts []uint64, tokens [][]byte, dsts []string, dstsAmounts []uint64, dstsTokens [][]byte, fee, computeFeeBlockHeight uint64) (tx *transaction.Transaction, err error) {

	if len(keys) != len(amounts) || len(amounts) != len(tokens) || len(amounts) == 0 {
		err = errors.New("Input lengths are a mismatch")
		return
	}
	if len(dsts) != len(dstsAmounts) {
		err = errors.New("Output lengths are a mismatch")
		return
	}

	var privateKeys []addresses.PrivateKey

	var vin []transaction_simple.TransactionSimpleInput
	for i := 0; i < len(keys); i++ {

		privateKeys = append(privateKeys, addresses.PrivateKey{Key: keys[i]})

		var publicKey [33]byte
		if publicKey, err = privateKeys[i].GeneratePublicKey(); err != nil {
			return
		}

		vin = append(vin, transaction_simple.TransactionSimpleInput{
			Amount:    amounts[i],
			PublicKey: publicKey,
			Token:     tokens[i],
		})
	}

	var vout []transaction_simple.TransactionSimpleOutput
	for i := 0; i < len(dsts); i++ {

		var outAddress *addresses.Address
		if outAddress, err = addresses.DecodeAddr(dsts[i]); err != nil {
			return
		}

		var publicKeyHash [20]byte
		switch outAddress.Version {
		case addresses.SimplePublicKeyHash:
			publicKeyHash = *helpers.Byte20(outAddress.PublicKey)
		case addresses.SimplePublicKey:
			publicKeyHash = crypto.ComputePublicKeyHash(*helpers.Byte33(outAddress.PublicKey))
		}

		vout = append(vout, transaction_simple.TransactionSimpleOutput{
			PublicKeyHash: publicKeyHash,
			Amount:        dstsAmounts[i],
			Token:         dstsTokens[i],
		})
	}

	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TransactionTypeSimple,
		TxBase: transaction_simple.TransactionSimple{
			Nonce: nonce,
			Vin:   vin,
			Vout:  vout,
		},
	}

	if computeFeeBlockHeight > 0 && fee == 0 {
		oldFee := uint64(1)
		for oldFee != fee {
			fee = fees.ComputeTxFees(uint64(len(tx.Serialize(true))), computeFeeBlockHeight)
			oldFee = fee
			tx.TxBase.(transaction_simple.TransactionSimple).Vin[0].Amount = amounts[0] + fee
		}
	}

	hash := tx.SerializeForSigning()
	for i, privateKey := range privateKeys {

		var signature [65]byte
		if signature, err = privateKey.Sign(&hash); err != nil {
			return
		}

		tx.TxBase.(transaction_simple.TransactionSimple).Vin[i].Signature = signature

	}

	return
}

func CreateUnstakeTx(nonce uint64, key [32]byte, unstakeAmount, fee, computeFeeBlockHeight uint64) (tx *transaction.Transaction, err error) {

	privateKey := addresses.PrivateKey{Key: key}
	var publicKey [33]byte
	if publicKey, err = privateKey.GeneratePublicKey(); err != nil {
		return
	}

	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TransactionTypeSimpleUnstake,
		TxBase: transaction_simple.TransactionSimple{
			Nonce: nonce,
			Extra: transaction_simple_unstake.TransactionSimpleUnstake{
				UnstakeAmount: unstakeAmount,
			},
			Vin: []transaction_simple.TransactionSimpleInput{
				{
					Amount:    fee,
					PublicKey: publicKey,
				},
			},
		},
	}

	if computeFeeBlockHeight > 0 && fee == 0 {
		oldFee := uint64(1)
		for oldFee != fee {
			fee = fees.ComputeTxFees(uint64(len(tx.Serialize(true))), computeFeeBlockHeight)
			oldFee = fee
			tx.TxBase.(transaction_simple.TransactionSimple).Vin[0].Amount = fee
		}
	}

	hash := tx.SerializeForSigning()
	var signature [65]byte

	if signature, err = privateKey.Sign(&hash); err != nil {
		return
	}

	tx.TxBase.(transaction_simple.TransactionSimple).Vin[0].Signature = signature

	return
}
