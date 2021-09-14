package transaction

import (
	"errors"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionBloom struct {
	Serialized helpers.HexBytes
	Size       uint64
	Hash       helpers.HexBytes
	HashStr    string
	bloomed    bool
}

func (tx *Transaction) BloomNow() {

	if tx.Bloom != nil {
		return
	}

	bloom := new(TransactionBloom)
	bloom.Serialized = tx.SerializeManualToBytes()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.Hash = cryptography.SHA3(bloom.Serialized)
	bloom.HashStr = string(bloom.Hash)
	bloom.bloomed = true
	tx.Bloom = bloom
}

func (tx *Transaction) BloomAll() (err error) {
	tx.BloomNow()
	return tx.BloomExtraNow()
}

func (tx *Transaction) BloomExtraNow() (err error) {
	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		serialized := tx.SerializeForSigning()
		err = tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).BloomNow(serialized)
	case transaction_type.TX_ZETHER:
		serialized := tx.SerializeForSigning()
		err = tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).BloomNow(serialized)
	default:
		err = errors.New("Invalid TxType")
	}
	return
}

func (tx *Transaction) BloomExtraVerified() (err error) {
	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		err = tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).BloomNowSignatureVerified()
	default:
		err = errors.New("Invalid TxType")
	}
	return
}

func (tx *Transaction) VerifyBloomAll() (err error) {
	if err = tx.Bloom.verifyIfBloomed(); err != nil {
		return
	}
	return tx.TransactionBaseInterface.VerifyBloomAll()
}

func (bloom *TransactionBloom) verifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("Tx is not bloomed")
	}
	return nil
}
