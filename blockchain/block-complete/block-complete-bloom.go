package block_complete

import (
	"bytes"
	"errors"
	"pandora-pay/helpers"
)

type BlockCompleteBloom struct {
	Serialized                helpers.HexBytes `json:"-"`
	Size                      uint64           `json:"size"`
	merkleTreeVerified        bool
	bloomedSize               bool
	bloomedMerkleTreeVerified bool
}

func (blkComplete *BlockComplete) BloomAll() (err error) {

	for _, tx := range blkComplete.Txs {
		if err = tx.BloomAll(); err != nil {
			return
		}
	}

	if err = blkComplete.Block.BloomNow(); err != nil {
		return
	}
	if err = blkComplete.BloomNow(); err != nil {
		return
	}

	return
}

func (blkComplete *BlockComplete) BloomNow() error {

	if blkComplete.Bloom == nil {
		blkComplete.Bloom = new(BlockCompleteBloom)
	}

	if !blkComplete.Bloom.bloomedSize {
		blkComplete.Bloom.Serialized = blkComplete.SerializeToBytes()
		blkComplete.Bloom.Size = uint64(len(blkComplete.Bloom.Serialized))
		blkComplete.Bloom.bloomedSize = true
	}
	if !blkComplete.Bloom.bloomedMerkleTreeVerified {
		blkComplete.Bloom.merkleTreeVerified = bytes.Equal(blkComplete.MerkleHash(), blkComplete.Block.MerkleHash)
		if !blkComplete.Bloom.merkleTreeVerified {
			return errors.New("Verify Merkle Hash failed")
		}
		blkComplete.Bloom.bloomedMerkleTreeVerified = true
	}

	return nil
}

func (blkComplete *BlockComplete) VerifyBloomAll() error {
	return blkComplete.Bloom.verifyIfBloomed()
}

func (bloom *BlockCompleteBloom) verifyIfBloomed() error {
	if !bloom.bloomedSize || !bloom.bloomedMerkleTreeVerified {
		return errors.New("block complete was not bloomed")
	}
	if bloom.Serialized == nil {
		return errors.New("block complete serialized was not read")
	}
	if bloom.Size == 0 {
		return errors.New("block complete size was not bloomed")
	}
	if !bloom.merkleTreeVerified {
		return errors.New("Verify Merkle Hash failed")
	}
	return nil
}
