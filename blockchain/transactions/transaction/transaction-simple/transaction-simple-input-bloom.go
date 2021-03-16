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
	vin.Bloom = new(TransactionSimpleInputBloom)
	vin.Bloom.bloomPublicKey(hashForSignature, vin.Signature)
}

func (bloom *TransactionSimpleInputBloom) bloomPublicKey(hashSerializedForSignature []byte, signature []byte) {
	publicKey, err := ecdsa.EcrecoverCompressed(hashSerializedForSignature, signature)
	if err != nil {
		panic(err)
	}
	bloom.PublicKey = publicKey
	bloom.PublicKeyHash = cryptography.ComputePublicKeyHash(bloom.PublicKey)
	bloom.bloomed = true
}

func (bloom *TransactionSimpleInputBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("vin is not bloomed")
	}
}
