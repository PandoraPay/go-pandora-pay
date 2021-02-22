package block

import (
	"bytes"
	"encoding/binary"
	"pandora-pay/helpers"
)

type BlockHeader struct {
	Version   uint64
	Timestamp uint64
	Height    uint64
}

func (blockHeader *BlockHeader) Serialize() []byte {

	var serialised bytes.Buffer

	buf := make([]byte, binary.MaxVarintLen64)
	n := binary.PutUvarint(buf, blockHeader.Version)
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Timestamp)
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Height)
	serialised.Write(buf[:n])

	return serialised.Bytes()
}

func (blockHeader *BlockHeader) Deserialize(buf []byte) (out []byte, err error) {

	out = buf

	blockHeader.Version, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	blockHeader.Timestamp, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	blockHeader.Height, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}

	return
}
