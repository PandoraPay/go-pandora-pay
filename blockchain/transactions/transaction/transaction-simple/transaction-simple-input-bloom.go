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

func (vin *TransactionSimpleInput) BloomNow(hash []byte) {
	vin.Bloom = new(TransactionSimpleInputBloom)
	vin.Bloom.bloomPublicKey(hash, vin.Signature)
}

func (bloom *TransactionSimpleInputBloom) bloomPublicKey(hash []byte, signature []byte) {
	out, err := ecdsa.Ecrecover(hash, signature)
	if err != nil {
		panic(err)
	}
	bloom.PublicKey = out
	bloom.PublicKey = cryptography.ComputePublicKeyHash(bloom.PublicKey)
	bloom.bloomed = true
}

func (bloom *TransactionSimpleInputBloom) VerifyIfBloomed() {
	if !bloom.bloomed {
		panic("vin is not bloomed")
	}
}
