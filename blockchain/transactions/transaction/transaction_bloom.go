package transaction

import (
	"errors"
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

func (tx *Transaction) bloomExtraNow() error {
	return tx.TransactionBaseInterface.BloomNow()
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
