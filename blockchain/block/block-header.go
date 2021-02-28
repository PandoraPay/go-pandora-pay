package block

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type BlockHeader struct {
	Version uint64
	Height  uint64
}

func (blockHeader *BlockHeader) Serialize(serialized *bytes.Buffer, temp []byte) {

	n := binary.PutUvarint(temp, blockHeader.Version)
	serialized.Write(temp[:n])

	n = binary.PutUvarint(temp, blockHeader.Height)
	serialized.Write(temp[:n])

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
