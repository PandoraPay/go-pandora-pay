package transaction

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
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

func (tx *Transaction) BloomAll() (err error) {

	if tx.Bloom != nil {
		return tx.bloomExtraNow()
	}

	if err = tx.validate(); err != nil {
		return
	}

	bloom := new(TransactionBloom)
	bloom.Serialized = tx.SerializeManualToBytes()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.Hash = cryptography.SHA3(bloom.Serialized)
	bloom.HashStr = string(bloom.Hash)
	bloom.bloomed = true
	tx.Bloom = bloom

	return tx.bloomExtraNow()
}

func (tx *Transaction) bloomExtraNow() (err error) {
	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
		if base.Bloom != nil {
			return
		}
		err = base.BloomNow(tx.SerializeForSigning())
	case transaction_type.TX_ZETHER:
		base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
		if base.Bloom != nil {
			return
		}
		err = base.BloomNow(tx.SerializeForSigning())
	default:
		err = errors.New("Invalid TxType")
	}
	return
}

func (tx *Transaction) BloomExtraVerified() (err error) {
	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		err = tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).BloomNowSignatureVerified()
	case transaction_type.TX_ZETHER:
		err = tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).BloomNowSignatureVerified()
	default:
		err = errors.New("Invalid TxType")
	}
	return
}

func (tx *Transaction) VerifyBloomAll() (err error) {
	if tx.Bloom == nil {
		return errors.New("Tx was not bloomed")
	}
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
