package block

import "pandora-pay/cryptography/ecdsa"

type BlockBloom struct {
	Hash             []byte
	KernelHash       []byte
	hashForSignature []byte
	PublicKey        []byte
	bloomed          bool
}

func (blk *Block) BloomNow() {
	bloom := new(BlockBloom)

	bloom.Hash = blk.ComputeHash()
	bloom.KernelHash = blk.ComputeKernelHash()
	bloom.hashForSignature = blk.SerializeForSigning()
	publicKey, err := ecdsa.EcrecoverCompressed(bloom.hashForSignature, blk.Signature)
	if err != nil {
		panic(err)
	}
	bloom.PublicKey = publicKey
	bloom.bloomed = true

	blk.Bloom = bloom
}

func (blk *Block) VerifyBloomAll() {
	blk.Bloom.verifyIfBloomed()
}

func (bloom *BlockBloom) verifyIfBloomed() {
	if !bloom.bloomed {
		panic("Bloom was not ")
	}
}
