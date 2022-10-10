package block

import (
	"errors"
	"pandora-pay/helpers/advanced_buffers"
)

type BlockHeader struct {
	Version uint64 `json:"version" msgpack:"version"`
	Height  uint64 `json:"height" msgpack:"height"`
}

func (blockHeader *BlockHeader) validate() error {
	if blockHeader.Version != 0 {
		return errors.New("Invalid Block")
	}
	return nil
}

func (blockHeader *BlockHeader) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(blockHeader.Version)
	w.WriteUvarint(blockHeader.Height)
}

func (blockHeader *BlockHeader) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	if blockHeader.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if blockHeader.Height, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
