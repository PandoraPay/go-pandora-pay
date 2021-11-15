package min_max_heap

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type HeapElement struct {
	hash_map.HashMapElementSerializableInterface
	Key   []byte
	Score float64
}

type HeapDictElement struct {
	hash_map.HashMapElementSerializableInterface
	Index uint64
}

func (self *HeapElement) Validate() error {
	return nil
}

func (self *HeapElement) Serialize(w *helpers.BufferWriter) {
	w.Write(self.Key)
	w.WriteFloat64(self.Score)
}

func (self *HeapElement) Deserialize(r *helpers.BufferReader) (err error) {
	if self.Key, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	self.Score, err = r.ReadFloat64()
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
