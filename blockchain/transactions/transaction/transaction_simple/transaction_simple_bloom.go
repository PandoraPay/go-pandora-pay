package transaction_simple

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleBloom struct {
	VinPublicKeyHashes [][]byte
	TransferMap        map[string]uint64
	bloomed            bool
}

func (tx *TransactionSimple) BloomNow() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionSimpleBloom)
	tx.Bloom.VinPublicKeyHashes = make([][]byte, len(tx.Vin))

	amounts := make(map[string]uint64)

	for i, vin := range tx.Vin {
		tx.Bloom.VinPublicKeyHashes[i] = cryptography.GetPublicKeyHash(vin.PublicKey)

		sum := amounts[string(vin.Asset)]
		if err = helpers.SafeUint64Add(&sum, vin.Amount); err != nil {
			return
		}
		amounts[string(vin.Asset)] = sum
	}

	for _, vout := range tx.Vout {
		sum := amounts[string(vout.Asset)]
		if err = helpers.SafeUint64Sub(&sum, vout.Amount); err != nil {
			return
		}
		amounts[string(vout.Asset)] = sum
	}

	tx.Bloom.TransferMap = amounts
	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	return nil
}
