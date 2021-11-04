package min_heap

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type MinHeapElement struct {
	helpers.SerializableInterface
	Key   []byte
	Score uint64
}

type MinHeapDictElement struct {
	helpers.SerializableInterface
	Index uint64
}

func (self *MinHeapElement) Serialize(w *helpers.BufferWriter) {
	w.Write(self.Key)
	w.WriteUvarint(self.Score)
}

func (self *MinHeapElement) Deserialize(r *helpers.BufferReader) (err error) {
	if self.Key, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	self.Score, err = r.ReadUvarint()
	return
}

func (self *MinHeapDictElement) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(self.Index)
}

func (self *MinHeapDictElement) Deserialize(r *helpers.BufferReader) (err error) {
	self.Index, err = r.ReadUvarint()
	return
}
