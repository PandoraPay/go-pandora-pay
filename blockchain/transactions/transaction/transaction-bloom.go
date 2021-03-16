package transaction

import (
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
)

type TransactionBloom struct {
	Serialized []byte
	Size       uint64
	Hash       []byte
	HashStr    string
	bloomed    bool
}

func (tx *Transaction) BloomNow() {
	tx.Bloom = new(TransactionBloom)
	tx.Bloom.BloomTransactionHash(tx)
}

func (tx *Transaction) BloomAll() {
	tx.BloomNow()
	tx.BloomExtraNow(false)
}

func (tx *Transaction) BloomExtraNow(signaturedWasVerifiedBefore bool) {
	switch tx.TxType {
	case transaction_type.TxSimple:
		base := tx.TxBase.(*transaction_simple.TransactionSimple)
		base.BloomNow(tx.SerializeForSigning(), signaturedWasVerifiedBefore)
	}
}

func (tx *Transaction) VerifyBloomAll() {
	tx.Bloom.VerifyIfBloomed()
	switch tx.TxType {
	case transaction_type.TxSimple:
		base := tx.TxBase.(*transaction_simple.TransactionSimple)
		base.VerifyBloomAll()
	default:
		panic("invalid tx.TxType")
	}
}

func (bloom *TransactionBloom) BloomTransactionHash(tx *Transaction) {
	bloom.Serialized = tx.Serialize()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.Hash = cryptography.SHA3Hash(bloom.Serialized)
	bloom.HashStr = string(bloom.Hash)
	bloom.bloomed = true
}

func (bloom *TransactionBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("Tx is not bloomed")
	}
}
