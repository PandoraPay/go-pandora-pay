package block

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
)

type BlockBloom struct {
	Serialized                 helpers.HexBytes `json:"-"`
	Hash                       helpers.HexBytes `json:"hash"`
	KernelHash                 helpers.HexBytes `json:"kernelHash"`
	DelegatedPublicKey         helpers.HexBytes `json:"delegatedPublicKey"`
	DelegatedSignatureVerified bool             `json:"delegatedSignatureVerified"`
	bloomedHash                bool             `json:"-"`
	bloomedKernelHash          bool             `json:"-"`
}

func (blk *Block) BloomSerializedNow(serialized []byte) {
	blk.Bloom = &BlockBloom{
		Serialized:  serialized,
		Hash:        cryptography.SHA3Hash(serialized),
		bloomedHash: true,
	}
}

func (blk *Block) BloomNow() (err error) {

	if blk.Bloom == nil {
		blk.Bloom = new(BlockBloom)
	}

	if !blk.Bloom.bloomedHash {
		blk.Bloom.Serialized = blk.SerializeManualToBytes()
		blk.Bloom.Hash = cryptography.SHA3Hash(blk.Bloom.Serialized)
		blk.Bloom.bloomedHash = true
	}
	if !blk.Bloom.bloomedKernelHash {

		blk.Bloom.KernelHash = blk.ComputeKernelHash()
		hashForSignature := blk.SerializeForSigning()

		var delegatedPublicKey []byte
		if delegatedPublicKey, err = ecdsa.EcrecoverCompressed(hashForSignature, blk.Signature); err != nil {
			return
		}

		blk.Bloom.DelegatedPublicKey = delegatedPublicKey
		blk.Bloom.DelegatedSignatureVerified = ecdsa.VerifySignature(delegatedPublicKey, hashForSignature, blk.Signature[0:64])
		if !blk.Bloom.DelegatedSignatureVerified {
			return errors.New("BLock signature is invalid")
		}

		blk.Bloom.bloomedKernelHash = true
	}

	return
}

func (bloom *BlockBloom) verifyIfBloomed() error {
	if !bloom.bloomedHash {
		return errors.New("Bloom was not validated")
	}
	if !bloom.bloomedKernelHash {
		return errors.New("Bloom KernelHash was not validated")
	}

	return nil
}
