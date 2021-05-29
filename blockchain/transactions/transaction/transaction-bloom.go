package transaction

import (
	"errors"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionBloom struct {
	Serialized helpers.HexBytes `json:"serialized"`
	Size       uint64           `json:"size"`
	Hash       helpers.HexBytes `json:"hash"`
	HashStr    string           `json:"-"`
	bloomed    bool             `json:"bloomed"`
}

func (tx *Transaction) BloomNow() {

	if tx.Bloom != nil {
		return
	}

	bloom := new(TransactionBloom)
	bloom.Serialized = tx.SerializeToBytes()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.Hash = cryptography.SHA3Hash(bloom.Serialized)
	bloom.HashStr = string(bloom.Hash)
	bloom.bloomed = true
	tx.Bloom = bloom
}

func (tx *Transaction) BloomAll() (err error) {
	tx.BloomNow()
	return tx.BloomExtraNow(false)
}

func (tx *Transaction) BloomExtraNow(signatureWasVerifiedBefore bool) (err error) {
	switch tx.TxType {
	case transaction_type.TX_SIMPLE:
		base := tx.TxBase.(*transaction_simple.TransactionSimple)
		if err = base.BloomNow(tx.SerializeForSigning(), signatureWasVerifiedBefore); err != nil {
			return
		}
	}
	return
}

func (tx *Transaction) VerifyBloomAll() (err error) {
	if err = tx.Bloom.verifyIfBloomed(); err != nil {
		return
	}
	return tx.TxBase.VerifyBloomAll()
}

func (bloom *TransactionBloom) verifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("Tx is not bloomed")
	}
	return nil
}
