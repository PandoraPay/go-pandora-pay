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

func (blockHeader *BlockHeader) Deserialize(buf []byte) (out []byte, err error) {

	blockHeader.Version, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	blockHeader.Height, buf, err = helpers.DeserializeNumber(buf)
	if err != nil {
		return
	}

	out = buf
	return
}
