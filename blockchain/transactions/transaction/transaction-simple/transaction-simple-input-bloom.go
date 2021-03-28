package transaction_simple

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
)

type TransactionSimpleInputBloom struct {
	PublicKey     []byte //30
	PublicKeyHash []byte //20
	bloomed       bool
}

func (vin *TransactionSimpleInput) BloomNow(hashForSignature []byte) (err error) {

	if vin.Bloom != nil {
		return
	}

	bloom := new(TransactionSimpleInputBloom)

	if bloom.PublicKey, err = ecdsa.EcrecoverCompressed(hashForSignature, vin.Signature); err != nil {
		return
	}

	bloom.PublicKeyHash = cryptography.ComputePublicKeyHash(bloom.PublicKey)
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
