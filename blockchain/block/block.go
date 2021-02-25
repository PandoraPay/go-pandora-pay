package block

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/crypto"
	"pandora-pay/crypto/ecdsa"
	"pandora-pay/helpers"
)

type Block struct {
	BlockHeader
	MerkleHash crypto.Hash

	PrevHash       crypto.Hash
	PrevKernelHash crypto.Hash

	Timestamp uint64

	Forger    [33]byte // 33 byte public key
	Signature [65]byte // 65 byte signature
}

func (blk *Block) ComputeHash() crypto.Hash {
	return crypto.SHA3Hash(blk.Serialize())
}

func (blk *Block) ComputeKernelHash() crypto.Hash {
	return crypto.SHA3Hash(blk.SerializeBlock(false, false, true, true, false))
}

func (blk *Block) SerializeForSigning() crypto.Hash {
	return crypto.SHA3Hash(blk.SerializeBlock(true, true, true, true, false))
}

func (blk *Block) VerifySignature() bool {
	hash := blk.SerializeForSigning()
	return ecdsa.VerifySignature(blk.Forger[:], hash[:], blk.Signature[0:64])
}

func (blk *Block) SerializeBlock(inclMerkleHash bool, inclPrevHash bool, inclTimestamp bool, inclForger bool, inclSignature bool) []byte {

	var serialized bytes.Buffer
	temp := make([]byte, binary.MaxVarintLen64)

	blk.BlockHeader.Serialize(&serialized, temp)

	if inclMerkleHash {
		serialized.Write(blk.MerkleHash[:])
	}

	if inclPrevHash {
		serialized.Write(blk.PrevHash[:])
	}

	serialized.Write(blk.PrevKernelHash[:])

	if inclTimestamp {
		n := binary.PutUvarint(temp, blk.Timestamp)
		serialized.Write(temp[:n])
	}

	if inclForger {
		serialized.Write(blk.Forger[:])
	}

	if inclSignature {
		serialized.Write(blk.Signature[:])
	}

	return serialized.Bytes()
}

func (blk *Block) Serialize() []byte {
	return blk.SerializeBlock(true, true, true, true, true)
}

func (blk *Block) Deserialize(buf []byte) (out []byte, err error) {

	buf, err = blk.BlockHeader.Deserialize(buf)
	if err != nil {
		return
	}

	blk.MerkleHash, buf, err = helpers.DeserializeHash(buf, crypto.HashSize)
	if err != nil {
		return
	}

	blk.PrevHash, buf, err = helpers.DeserializeHash(buf, crypto.HashSize)
	if err != nil {
		return
	}

	blk.PrevKernelHash, buf, err = helpers.DeserializeHash(buf, crypto.HashSize)
	if err != nil {
		return
	}

	blk.Timestamp, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	var data []byte
	data, buf, err = helpers.DeserializeBuffer(buf, 33)
	if err != nil {
		return
	}
	copy(blk.Forger[:], data)

	data, buf, err = helpers.DeserializeBuffer(buf, 65)
	if err != nil {
		return
	}
	copy(blk.Signature[:], data)

	out = buf
	return
}
