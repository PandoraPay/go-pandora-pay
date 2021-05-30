package block_complete

import (
	"bytes"
	"errors"
	"pandora-pay/helpers"
)

type BlockCompleteBloom struct {
	Serialized         helpers.HexBytes `json:"serialized"`
	merkleTreeVerified bool
	Size               uint64
	bloomed            bool
}

func (blkComplete *BlockComplete) BloomNow() error {

	if blkComplete.Bloom != nil {
		return nil
	}

	bloom := new(BlockCompleteBloom)

	bloom.merkleTreeVerified = bytes.Equal(blkComplete.MerkleHash(), blkComplete.Block.MerkleHash)
	if !bloom.merkleTreeVerified {
		return errors.New("Verify Merkle Hash failed")
	}

	bloom.Serialized = blkComplete.SerializeToBytes()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.bloomed = true
	blkComplete.Bloom = bloom
	return nil
}

func (blkComplete *BlockComplete) VerifyBloomAll() error {
	return blkComplete.Bloom.verifyIfBloomed()
}

func (bloom *BlockCompleteBloom) verifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("block complete was not bloomed")
	}
	if bloom.Size == 0 {
		return errors.New("block complete size was not bloomed")
	}
	if !bloom.merkleTreeVerified {
		return errors.New("Verify Merkle Hash failed")
	}
	return nil
}
