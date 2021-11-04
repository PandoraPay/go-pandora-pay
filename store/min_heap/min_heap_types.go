package min_heap

import "pandora-pay/helpers"

type MinHeapElement struct {
	helpers.SerializableInterface
	Data  helpers.SerializableInterface
	Score uint64
}

func (self *MinHeapElement) Serialize(w *helpers.BufferWriter) {
	self.Data.Serialize(w)
	w.WriteUvarint(self.Score)
}

func (self *MinHeapElement) Deserialize(r *helpers.BufferReader) (err error) {
	if err = self.Data.Deserialize(r); err != nil {
		return
	}
	self.Score, err = r.ReadUvarint()
	return
}
