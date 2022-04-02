package transaction_simple

import (
	"errors"
	"pandora-pay/cryptography"
)

type TransactionSimpleBloom struct {
	VinPublicKeyHashes [][]byte
	bloomed            bool
}

func (tx *TransactionSimple) BloomNow() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionSimpleBloom)
	tx.Bloom.VinPublicKeyHashes = make([][]byte, len(tx.Vin))
	for i, vin := range tx.Vin {
		tx.Bloom.VinPublicKeyHashes[i] = cryptography.GetPublicKeyHash(vin.PublicKey)
	}

	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	return nil
}
