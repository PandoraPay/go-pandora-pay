package block

import (
	"errors"
	"pandora-pay/helpers"
)

type BlockHeader struct {
	helpers.SerializableInterface
	Version uint64
	Height  uint64
}

func (blockHeader *BlockHeader) Validate() error {
	if blockHeader.Version != 0 {
		return errors.New("Invalid Block")
	}
	return nil
}

func (blockHeader *BlockHeader) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(blockHeader.Version)
	writer.WriteUvarint(blockHeader.Height)
}

func (blockHeader *BlockHeader) Deserialize(reader *helpers.BufferReader) (err error) {
	if blockHeader.Version, err = reader.ReadUvarint(); err != nil {
		return
	}
	if blockHeader.Height, err = reader.ReadUvarint(); err != nil {
		return
	}
	return
}
