package block

import (
	"bytes"
	"pandora-pay/crypto"
	"pandora-pay/crypto/ecdsa"
	"pandora-pay/helpers"
)

type Block struct {
	BlockHeader
	MerkleHash crypto.Hash

	PrevHash       crypto.Hash
	PrevKernelHash crypto.Hash

	Forger    []byte // 33 byte public key
	Signature []byte // 65 byte signature
}

type BlockComplete struct {
	block *Block
	//txs []*transaction
}

func (block *Block) ComputeHash() crypto.Hash {
	buf := block.Serialize()
	return crypto.SHA3Hash(buf)
}

func (block *Block) ComputeKernelHash() crypto.Hash {
	return crypto.SHA3Hash(block.serializeBlock(false, false, false))
}

func (block *Block) SerializeForSigning() crypto.Hash {
	return crypto.SHA3Hash(block.serializeBlock(true, true, false))
}

func (block *Block) VerifySignature() bool {
	hash := block.SerializeForSigning()
	return ecdsa.VerifySignature(block.Forger, hash[:], block.Signature[0:64])
}

func (block *Block) serializeBlock(inclMerkleHash bool, inclPrevHash bool, inclSignature bool) []byte {
	var serialized bytes.Buffer

	serialized.Write(block.BlockHeader.Serialize())

	if inclMerkleHash {
		serialized.Write(block.MerkleHash[:])
	}

	if inclPrevHash {
		serialized.Write(block.PrevHash[:])
	}

	serialized.Write(block.PrevKernelHash[:])

	serialized.Write(block.Forger[:])

	if inclSignature {
		serialized.Write(block.Signature[:])
	}

	return serialized.Bytes()
}

func (block *Block) Serialize() []byte {
	return block.serializeBlock(true, true, true)
}

func (block *Block) Deserialize(buf []byte) (out []byte, err error) {

	out = buf

	out, err = block.BlockHeader.Deserialize(out)
	if err != nil {
		return
	}

	var hash []byte
	hash, out, err = helpers.DeserializeBuffer(out, crypto.HashSize)
	if err != nil {
		return
	}
	copy(block.MerkleHash[:], hash)

	hash, out, err = helpers.DeserializeBuffer(out, crypto.HashSize)
	if err != nil {
		return
	}
	copy(block.PrevHash[:], hash)

	hash, out, err = helpers.DeserializeBuffer(out, crypto.HashSize)
	if err != nil {
		return
	}
	copy(block.PrevKernelHash[:], hash)

	block.Forger, out, err = helpers.DeserializeBuffer(out, 33)
	if err != nil {
		return
	}

	block.Signature, out, err = helpers.DeserializeBuffer(out, 65)
	if err != nil {
		return
	}

	return
}
