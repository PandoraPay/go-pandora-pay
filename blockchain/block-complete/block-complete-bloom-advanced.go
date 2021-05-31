package block_complete

import (
	"bytes"
	"errors"
)

type BlockCompleteBloomAdvanced struct {
	merkleTreeVerified bool `json:"-"`
	bloomed            bool `json:"-"`
}

func (blkComplete *BlockComplete) BloomAdvancedNow() error {

	if blkComplete.BloomAdvanced != nil {
		return nil
	}

	bloom := new(BlockCompleteBloomAdvanced)

	bloom.merkleTreeVerified = bytes.Equal(blkComplete.MerkleHash(), blkComplete.Block.MerkleHash)
	if !bloom.merkleTreeVerified {
		return errors.New("Verify Merkle Hash failed")
	}

	bloom.bloomed = true
	blkComplete.BloomAdvanced = bloom
	return nil
}

func (bloom *BlockCompleteBloomAdvanced) verifyIfBloomedAdvanced() error {
	if !bloom.bloomed {
		return errors.New("block complete advanced was not bloomed")
	}
	if !bloom.merkleTreeVerified {
		return errors.New("Verify Merkle Hash failed")
	}
	return nil
}
