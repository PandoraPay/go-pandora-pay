package block

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/advanced_buffers"
)

type Block struct {
	*BlockHeader
	MerkleHash     []byte      `json:"merkleHash" msgpack:"merkleHash"`          //32 byte
	PrevHash       []byte      `json:"prevHash"  msgpack:"prevHash"`             //32 byte
	PrevKernelHash []byte      `json:"prevKernelHash"  msgpack:"prevKernelHash"` //32 byte
	Timestamp      uint64      `json:"timestamp" msgpack:"timestamp"`
	StakingAmount  uint64      `json:"stakingAmount" msgpack:"stakingAmount"`
	StakingNonce   []byte      `json:"stakingNonce" msgpack:"stakingNonce"` // 33 byte public key can also be found into the accounts tree
	Bloom          *BlockBloom `json:"bloom" msgpack:"bloom"`
}

func CreateEmptyBlock() *Block {
	return &Block{
		BlockHeader: &BlockHeader{},
	}
}

func (blk *Block) Validate() error {
	if err := blk.BlockHeader.Validate(); err != nil {
		return err
	}

	return nil
}

func (blk *Block) Verify() error {
	return blk.Bloom.verifyIfBloomed()
}

func (blk *Block) computeHash() []byte {
	return cryptography.SHA3(helpers.SerializeToBytes(blk))
}

func (blk *Block) ComputeKernelHash() []byte {
	writer := advanced_buffers.NewBufferWriter()
	blk.AdvancedSerialization(writer, true, false)
	return cryptography.SHA3(writer.Bytes())
}

func (blk *Block) SerializeForSigning() []byte {
	writer := advanced_buffers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, false)
	return cryptography.SHA3(writer.Bytes())
}

func (blk *Block) AdvancedSerialization(w *advanced_buffers.BufferWriter, kernelHash bool, inclSignature bool) {

	blk.BlockHeader.Serialize(w)

	if !kernelHash {
		w.Write(blk.MerkleHash)
		w.Write(blk.PrevHash)
	}

	w.Write(blk.PrevKernelHash)

	if !kernelHash {
		w.WriteUvarint(blk.StakingAmount)
	}

	w.WriteUvarint(blk.Timestamp)

	w.Write(blk.StakingNonce)

}

func (blk *Block) SerializeForForging(w *advanced_buffers.BufferWriter) {
	blk.AdvancedSerialization(w, true, false)
}

func (blk *Block) Serialize(w *advanced_buffers.BufferWriter) {
	w.Write(blk.Bloom.Serialized)
}

func (blk *Block) SerializeManualToBytes() []byte {
	writer := advanced_buffers.NewBufferWriter()
	blk.AdvancedSerialization(writer, false, true)
	return writer.Bytes()
}

func (blk *Block) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	first := r.Position

	if err = blk.BlockHeader.Deserialize(r); err != nil {
		return
	}
	if blk.MerkleHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.PrevHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.PrevKernelHash, err = r.ReadHash(); err != nil {
		return
	}
	if blk.StakingAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.Timestamp, err = r.ReadUvarint(); err != nil {
		return
	}
	if blk.StakingNonce, err = r.ReadBytes(32); err != nil {
		return
	}

	serialized := r.Buf[first:r.Position]
	blk.BloomSerializedNow(serialized)

	return
}
