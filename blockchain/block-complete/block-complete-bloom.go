package block_complete

import "bytes"

type BlockCompleteBloom struct {
	merkleTreeVerified bool
	bloomed            bool
}

func (blkComplete *BlockComplete) BloomNow() {
	bloom := new(BlockCompleteBloom)
	bloom.merkleTreeVerified = bytes.Equal(blkComplete.MerkleHash(), blkComplete.Block.MerkleHash)
	if !bloom.merkleTreeVerified {
		panic("Verify Merkle Hash failed")
	}
	bloom.bloomed = true
	blkComplete.Bloom = bloom
}

func (blkComplete *BlockComplete) VerifyBloomAll() {
	blkComplete.Bloom.verifyIfBloomed()
}

func (bloom *BlockCompleteBloom) verifyIfBloomed() {
	if !bloom.bloomed {
		panic("block complete was not bloomed")
	}
}
