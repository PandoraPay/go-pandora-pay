package transaction

import (
	"errors"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
)

type Transaction struct {
	Version uint64
	TxType  transaction_type.TransactionType
	TxBase  interface{}
}

func (tx *Transaction) ComputeFees() (fees map[string]uint64, err error) {

	fees = make(map[string]uint64)

	switch tx.TxType {
	case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:
		base := tx.TxBase.(transaction_simple.TransactionSimple)
		err = base.ComputeFees(fees)
	default:
		err = errors.New("Invalid type")
	}
	return
}

func (tx *Transaction) SerializeForSigning() helpers.Hash {
	return crypto.SHA3Hash(tx.Serialize(false))
}

func (tx *Transaction) VerifySignature() bool {

	hash := tx.SerializeForSigning()
	switch tx.TxType {
	case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:
		base := tx.TxBase.(transaction_simple.TransactionSimple)
		return base.VerifySignature(hash)
	default:
		return false
	}

}

func (tx *Transaction) ComputeHash() helpers.Hash {
	return crypto.SHA3Hash(tx.Serialize(true))
}

func (tx *Transaction) Serialize(inclSignature bool) []byte {
	writer := helpers.NewBufferWriter()

	writer.WriteUvarint(tx.Version)
	writer.WriteUvarint(uint64(tx.TxType))

	switch tx.TxType {
	case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:
		base := tx.TxBase.(transaction_simple.TransactionSimple)
		base.Serialize(writer, inclSignature, tx.TxType)
	default:
	}

	return writer.Bytes()
}

func (tx *Transaction) Validate() (err error) {
	if tx.Version != 0 {
		return errors.New("Version is invalid")
	}
	if transaction_type.TransactionTypeEND < tx.TxType {
		return errors.New("VersionType is invalid")
	}

	switch tx.TxType {
	case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:
		base := tx.TxBase.(transaction_simple.TransactionSimple)
		if err = base.Validate(tx.TxType); err != nil {
			return
		}
	}

	return
}

func (tx *Transaction) Deserialize(buf []byte) (err error) {
	reader := helpers.NewBufferReader(buf)
	var n uint64

	if tx.Version, err = reader.ReadUvarint(); err != nil {
		return
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	tx.TxType = transaction_type.TransactionType(n)

	switch tx.TxType {
	case transaction_type.TransactionTypeSimple, transaction_type.TransactionTypeSimpleUnstake:
		base := transaction_simple.TransactionSimple{}
		if err = base.Deserialize(reader, tx.TxType); err != nil {
			return
		}
		tx.TxBase = base
	default:
		return errors.New("Transaction type is invalid")
	}

	return
}

func (tx *Transaction) IsTransactionSimple() bool {
	return tx.TxType == transaction_type.TransactionTypeSimple || tx.TxType == transaction_type.TransactionTypeSimpleUnstake
}
