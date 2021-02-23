package block

import (
	"bytes"
	"pandora-pay/crypto"
	"pandora-pay/helpers"
)

type Block struct {
	BlockHeader
	MerkleHash crypto.Hash

	Forger    []byte // 33 byte public key
	Signature []byte // 65 byte signature

}

type BlockComplete struct {
	block *Block
	//txs []*transaction
}

//func (block *Block) ComputeHash() (crypto.Hash, error) {
//
//}
//
//func (block *Block) ComputeKernelHash() (crypto.Hash, error) {
//
//}

func (block *Block) Serialize() []byte {

	var serialized bytes.Buffer

	header := block.BlockHeader.Serialize()
	serialized.Write(header)

	serialized.Write(block.MerkleHash[:])

	serialized.Write(block.Forger[:])

	serialized.Write(block.Signature[:])

	return serialized.Bytes()
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
