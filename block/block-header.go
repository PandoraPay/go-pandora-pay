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

func (blockHeader *BlockHeader) Serialize() []byte {

	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, blockHeader.Version)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Height)
	serialized.Write(buf[:n])

	return serialized.Bytes()
}

func (blockHeader *BlockHeader) Deserialize(buf []byte) (out []byte, err error) {

	out = buf

	blockHeader.Version, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	blockHeader.Height, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	return
}
