package block

import (
	"pandora-pay/helpers"
)

type BlockHeader struct {
	Version uint64
	Height  uint64
}

func (blockHeader *BlockHeader) Validate() {
	if blockHeader.Version != 0 {
		panic("Invalid Block")
	}
}

func (blockHeader *BlockHeader) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(blockHeader.Version)
	writer.WriteUvarint(blockHeader.Height)
}

func (blockHeader *BlockHeader) Deserialize(reader *helpers.BufferReader) {
	blockHeader.Version = reader.ReadUvarint()
	blockHeader.Height = reader.ReadUvarint()
	return
}
