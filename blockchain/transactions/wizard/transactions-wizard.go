package wizard

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction_simple_unstake"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
)

func CreateUnstake(nonce uint64, priv [32]byte, amount uint64) (tx *transaction.Transaction, err error) {

	privateKey := addresses.PrivateKey{Key: priv}
	var publicKey [33]byte
	if publicKey, err = privateKey.GeneratePublicKey(); err != nil {
		return
	}

	in := transaction_simple.TransactionSimpleInput{
		Amount:    amount,
		PublicKey: publicKey,
	}
	var vin []*transaction_simple.TransactionSimpleInput
	vin = append(vin, &in)

	tx = &transaction.Transaction{
		Version: 0,
		TxType:  transaction_type.TransactionTypeSimpleUnstake,
		TxBase: transaction_simple.TransactionSimple{
			Nonce: nonce,
			Extra: transaction_simple_unstake.TransactionSimpleUnstake{
				Fee: 0,
			},
			Vin: vin,
		},
	}

	hash := tx.SerializeForSigning()
	var signature [65]byte

	if signature, err = privateKey.Sign(&hash); err != nil {
		return
	}

	tx.TxBase.(transaction_simple.TransactionSimple).Vin[0].Signature = signature

	return
}
