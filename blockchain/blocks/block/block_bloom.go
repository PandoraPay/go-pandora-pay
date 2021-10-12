package block

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type BlockBloom struct {
	Serialized                 helpers.HexBytes `json:"-"`
	Hash                       helpers.HexBytes `json:"hash"`
	KernelHash                 helpers.HexBytes `json:"kernelHash"`
	DelegatedSignatureVerified bool             `json:"delegatedSignatureVerified"`
	bloomedHash                bool             `json:"-"`
	bloomedKernelHash          bool             `json:"-"`
}

func (blk *Block) BloomSerializedNow(serialized []byte) {
	blk.Bloom = &BlockBloom{
		Serialized:  serialized,
		Hash:        cryptography.SHA3(serialized),
		bloomedHash: true,
	}
}

func (blk *Block) BloomNow() (err error) {

	if err = blk.validate(); err != nil {
		return
	}

	if blk.Bloom == nil {
		blk.Bloom = new(BlockBloom)
	}

	if !blk.Bloom.bloomedHash {
		blk.Bloom.Serialized = blk.SerializeManualToBytes()
		blk.Bloom.Hash = cryptography.SHA3(blk.Bloom.Serialized)
		blk.Bloom.bloomedHash = true
	}
	if !blk.Bloom.bloomedKernelHash {

		blk.Bloom.KernelHash = blk.ComputeKernelHash()
		hashForSignature := blk.SerializeForSigning()

		blk.Bloom.DelegatedSignatureVerified = crypto.VerifySignature(hashForSignature, blk.Signature, blk.DelegatedStakePublicKey)
		if !blk.Bloom.DelegatedSignatureVerified {
			return errors.New("Block signature is invalid")
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
