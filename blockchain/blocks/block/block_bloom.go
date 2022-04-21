package block

import (
	"errors"
	"pandora-pay/cryptography"
)

type BlockBloom struct {
	Serialized        []byte `json:"-" msgpack:"-"`
	Hash              []byte `json:"hash" msgpack:"hash"`
	KernelHash        []byte `json:"kernelHash" msgpack:"kernelHash"`
	KernelHashStaked  []byte `json:"-" msgpack:"-"`
	SignatureVerified bool   `json:"signatureVerified" msgpack:"signatureVerified"`
	bloomedHash       bool
	bloomedKernelHash bool
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
		if blk.Bloom.KernelHashStaked, err = cryptography.ComputeKernelHash(blk.Bloom.KernelHash, blk.StakingAmount); err != nil {
			return
		}

		hashForSignature := blk.SerializeForSigning()

		blk.Bloom.SignatureVerified = cryptography.VerifySignature(blk.DelegatedStakePublicKey, hashForSignature, blk.Signature)
		if !blk.Bloom.SignatureVerified {
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
