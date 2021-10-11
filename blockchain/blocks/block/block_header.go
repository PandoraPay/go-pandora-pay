package block

import (
	"errors"
	"pandora-pay/helpers"
)

type BlockHeader struct {
	helpers.SerializableInterface `json:"-"`
	Version                       uint64 `json:"version"`
	Height                        uint64 `json:"height"`
}

func (blockHeader *BlockHeader) validate() error {
	if blockHeader.Version != 0 {
		return errors.New("Invalid Block")
	}
	return nil
}

func (blockHeader *BlockHeader) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(blockHeader.Version)
	w.WriteUvarint(blockHeader.Height)
}

func (blockHeader *BlockHeader) Deserialize(r *helpers.BufferReader) (err error) {
	if blockHeader.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if blockHeader.Height, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
