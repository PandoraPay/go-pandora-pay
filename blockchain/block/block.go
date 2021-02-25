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
	buf := make([]byte, binary.MaxVarintLen64)

	blk.BlockHeader.Serialize(&serialized, buf)

	if inclMerkleHash {
		serialized.Write(blk.MerkleHash[:])
	}

	if inclPrevHash {
		serialized.Write(blk.PrevHash[:])
	}

	serialized.Write(blk.PrevKernelHash[:])

	if inclTimestamp {
		n := binary.PutUvarint(buf, blk.Timestamp)
		serialized.Write(buf[:n])
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

	var hash []byte
	hash, buf, err = helpers.DeserializeBuffer(buf, crypto.HashSize)
	if err != nil {
		return
	}
	copy(blk.MerkleHash[:], hash)

	hash, buf, err = helpers.DeserializeBuffer(buf, crypto.HashSize)
	if err != nil {
		return
	}
	copy(blk.PrevHash[:], hash)

	hash, buf, err = helpers.DeserializeBuffer(buf, crypto.HashSize)
	if err != nil {
		return
	}
	copy(blk.PrevKernelHash[:], hash)

	blk.Timestamp, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	hash, buf, err = helpers.DeserializeBuffer(buf, 33)
	if err != nil {
		return
	}
	copy(blk.Forger[:], hash)

	hash, buf, err = helpers.DeserializeBuffer(buf, 65)
	if err != nil {
		return
	}
	copy(blk.Signature[:], hash)

	out = buf
	return
}
