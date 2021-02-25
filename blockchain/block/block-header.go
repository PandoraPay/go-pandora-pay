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

func (blockHeader *BlockHeader) Serialize(serialized *bytes.Buffer, buf []byte) {

	n := binary.PutUvarint(buf, blockHeader.Version)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Height)
	serialized.Write(buf[:n])

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
