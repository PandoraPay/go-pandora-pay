package block_complete

import (
	"errors"
	"pandora-pay/helpers"
)

type BlockCompleteBloom struct {
	Serialized helpers.HexBytes `json:"-"`
	Size       uint64           `json:"size"`
	bloomed    bool             `json:"-"`
}

func (blkComplete *BlockComplete) BloomNow() error {

	if blkComplete.Bloom != nil {
		return nil
	}

	bloom := new(BlockCompleteBloom)

	bloom.Serialized = blkComplete.SerializeToBytes()
	bloom.Size = uint64(len(bloom.Serialized))
	bloom.bloomed = true
	blkComplete.Bloom = bloom
	return nil
}

func (bloom *BlockCompleteBloom) verifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("block complete was not bloomed")
	}
	if bloom.Serialized == nil {
		return errors.New("block complete serialized was not read")
	}
	if bloom.Size == 0 {
		return errors.New("block complete size was not bloomed")
	}
	return nil
}
