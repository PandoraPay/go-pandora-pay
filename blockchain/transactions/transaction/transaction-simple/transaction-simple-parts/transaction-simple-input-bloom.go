package transaction_simple_parts

import (
	"errors"
)

type TransactionSimpleInputBloom struct {
	bloomed bool
}

func (vin *TransactionSimpleInput) BloomNow() (err error) {

	if vin.Bloom != nil {
		return
	}

	bloom := new(TransactionSimpleInputBloom)

	bloom.bloomed = true

	vin.Bloom = bloom
	return
}

func (bloom *TransactionSimpleInputBloom) VerifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("vin is not bloomed")
	}
	return nil
}
