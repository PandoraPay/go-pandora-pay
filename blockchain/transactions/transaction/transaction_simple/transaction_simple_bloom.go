package transaction_simple

import "errors"

type TransactionSimpleBloom struct {
	signatureVerified bool
	bloomed           bool
}

func (tx *TransactionSimple) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionSimpleBloom)

	tx.Bloom.signatureVerified = tx.VerifySignatureManually(hashForSignature)
	if !tx.Bloom.signatureVerified {
		return errors.New("Signature Failed for Transaction Simple")
	}

	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionSimple) BloomNowSignatureVerified() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionSimpleBloom)
	tx.Bloom.signatureVerified = true
	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	if !tx.signatureVerified {
		return errors.New("signatureVerified is false")
	}
	return nil
}
