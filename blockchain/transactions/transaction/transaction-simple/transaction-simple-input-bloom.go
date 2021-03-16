package transaction_simple

import (
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
)

type TransactionSimpleInputBloom struct {
	PublicKey     []byte //30
	PublicKeyHash []byte //20
	bloomed       bool
}

func (vin *TransactionSimpleInput) BloomNow(hashForSignature []byte) {
	bloom := new(TransactionSimpleInputBloom)

	publicKey, err := ecdsa.EcrecoverCompressed(hashForSignature, vin.Signature)
	if err != nil {
		panic(err)
	}

	bloom.PublicKey = publicKey
	bloom.PublicKeyHash = cryptography.ComputePublicKeyHash(bloom.PublicKey)
	bloom.bloomed = true

	vin.Bloom = bloom
}

func (bloom *TransactionSimpleInputBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("vin is not bloomed")
	}
}
