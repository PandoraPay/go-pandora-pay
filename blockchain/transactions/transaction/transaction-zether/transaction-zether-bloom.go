package transaction_zether

import "errors"

type TransactionZetherBloom struct {
	signatureVerified bool
	bloomed           bool
}

func (tx *TransactionZether) BloomNow(hashForSignature []byte) (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)

	tx.Bloom.signatureVerified = tx.VerifySignatureManually(hashForSignature)
	if !tx.Bloom.signatureVerified {
		return errors.New("Signature Failed for Transaction Simple")
	}

	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionZether) BloomNowSignatureVerified() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionZetherBloom)
	tx.Bloom.signatureVerified = true
	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionZetherBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	if !tx.signatureVerified {
		return errors.New("signatureVerified is false")
	}
	return nil
}
