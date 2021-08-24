package transaction_simple_parts

import (
	"errors"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type TransactionSimpleInputBloom struct {
	PublicKey helpers.HexBytes //32
	bloomed   bool
}

func (vin *TransactionSimpleInput) BloomNow(hashForSignature []byte) (err error) {

	if vin.Bloom != nil {
		return
	}

	bloom := new(TransactionSimpleInputBloom)

	if bloom.PublicKey, err = ecdsa.EcrecoverCompressed(hashForSignature, vin.Signature); err != nil {
		return
	}
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
