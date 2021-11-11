package min_max_heap

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type HeapElement struct {
	helpers.SerializableInterface
	Key   []byte
	Score uint64
}

type HeapDictElement struct {
	helpers.SerializableInterface
	Index uint64
}

func (self *HeapElement) Validate() error {
	return nil
}

func (self *HeapElement) Serialize(w *helpers.BufferWriter) {
	w.Write(self.Key)
	w.WriteUvarint(self.Score)
}

func (self *HeapElement) Deserialize(r *helpers.BufferReader) (err error) {
	if self.Key, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	self.Score, err = r.ReadUvarint()
	return
}

func (self *HeapDictElement) Validate() error {
	return nil
}

func (self *HeapDictElement) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(self.Index)
}

func (self *HeapDictElement) Deserialize(r *helpers.BufferReader) (err error) {
	self.Index, err = r.ReadUvarint()
	return
}
