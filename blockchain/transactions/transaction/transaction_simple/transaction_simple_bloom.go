package transaction_simple

import "errors"

type TransactionSimpleBloom struct {
	bloomed bool
}

func (tx *TransactionSimple) BloomNow() (err error) {

	if tx.Bloom != nil {
		return
	}

	tx.Bloom = new(TransactionSimpleBloom)

	tx.Bloom.bloomed = true
	return
}

func (tx *TransactionSimpleBloom) verifyIfBloomed() error {
	if !tx.bloomed {
		return errors.New("TransactionSimpleBloom was not bloomed")
	}
	return nil
}
