package block

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/helpers"
)

type BlockHeader struct {
	MajorVersion uint64
	MinorVersion uint64
	Timestamp    uint64
	Height       uint64
}

func (blockHeader *BlockHeader) Serialize() []byte {

	var serialized bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, blockHeader.MajorVersion)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.MinorVersion)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Timestamp)
	serialized.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Height)
	serialized.Write(buf[:n])

	return serialized.Bytes()
}

func (blockHeader *BlockHeader) Deserialize(buf []byte) (out []byte, err error) {

	out = buf

	blockHeader.MajorVersion, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}
	if blockHeader.MajorVersion != 0 {
		err = errors.New("MajorVersion is Invalid")
		return
	}

	blockHeader.MinorVersion, out, err = helpers.DeserializeNumber(out)
	if err != nil {
		return
	}
	if blockHeader.MinorVersion != 0 {
		err = errors.New("MinorVersion is Invalid")
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
