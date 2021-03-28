package block_complete

import (
	"bytes"
	"errors"
)

type BlockCompleteBloom struct {
	merkleTreeVerified bool
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
	return nil
}
