package block

import (
	"errors"
	"pandora-pay/cryptography/ecdsa"
)

type BlockBloom struct {
	Hash             []byte
	KernelHash       []byte
	hashForSignature []byte
	PublicKey        []byte
	bloomed          bool
}

func (blk *Block) BloomNow() (err error) {
	bloom := new(BlockBloom)

	bloom.Hash = blk.ComputeHash()
	bloom.KernelHash = blk.ComputeKernelHash()
	bloom.hashForSignature = blk.SerializeForSigning()
	if bloom.PublicKey, err = ecdsa.EcrecoverCompressed(bloom.hashForSignature, blk.Signature); err != nil {
		return
	}
	bloom.bloomed = true
	blk.Bloom = bloom
	return
}

func (blk *Block) VerifyBloomAll() error {
	return blk.Bloom.verifyIfBloomed()
}

func (bloom *BlockBloom) verifyIfBloomed() error {
	if !bloom.bloomed {
		return errors.New("Bloom was not validated")
	}
	return nil
}
