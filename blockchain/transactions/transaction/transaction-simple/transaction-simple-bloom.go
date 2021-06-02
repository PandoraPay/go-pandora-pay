package transaction_simple

import "errors"

type TransactionSimpleBloom struct {
	hashForSignature  []byte
	signatureVerified bool
	bloomed           bool
}

func (tx *TransactionSimple) BloomNow(hashForSignature []byte, signatureWasVerifiedBefore bool) (err error) {

	if tx.Bloom != nil {
		return
	}

	bloom := new(TransactionSimpleBloom)
	bloom.hashForSignature = hashForSignature

	for _, vin := range tx.Vin {
		if err = vin.BloomNow(hashForSignature); err != nil {
			return
		}
	}

	bloom.signatureVerified = signatureWasVerifiedBefore
	if !signatureWasVerifiedBefore {
		bloom.signatureVerified = tx.VerifySignatureManually(hashForSignature)
		if !bloom.signatureVerified {
			return errors.New("Signature Failed for Transaction Simple")
		}
	}
	bloom.bloomed = true
	tx.Bloom = bloom
	return nil
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	return nil
}
