package transaction_simple

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type TransactionSimpleInputBloom struct {
	PublicKey     helpers.HexBytes `json:"publicKey"`     //30
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash"` //20
	bloomed       bool             `json:"bloomed"`
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
