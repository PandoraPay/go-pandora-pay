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

	var serialised bytes.Buffer
	buf := make([]byte, binary.MaxVarintLen64)

	n := binary.PutUvarint(buf, blockHeader.MajorVersion)
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.MinorVersion)
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Timestamp)
	serialised.Write(buf[:n])

	n = binary.PutUvarint(buf, blockHeader.Height)
	serialised.Write(buf[:n])

	return serialised.Bytes()
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
