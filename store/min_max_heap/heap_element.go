package min_max_heap

import (
	"pandora-pay/helpers/advanced_buffers"
)

type HeapElement struct {
	HashmapKey []byte
	Key        []byte
	Score      float64
}

func (this *HeapElement) IsDeletable() bool {
	return false
}

func (this *HeapElement) SetKey(key []byte) {
	this.HashmapKey = key
}

func (this *HeapElement) SetIndex(index uint64) { //not indexable
}

func (this *HeapElement) GetIndex() uint64 { //not indexable
	return 0
}

func (this *HeapElement) Validate() error {
	return nil
}

func (this *HeapElement) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteByte(byte(len(this.Key)))
	w.Write(this.Key)
	w.WriteFloat64(this.Score)
}

func (this *HeapElement) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	var count byte
	if count, err = r.ReadByte(); err != nil {
		return
	}

	if this.Key, err = r.ReadBytes(int(count)); err != nil {
		return
	}
	this.Score, err = r.ReadFloat64()
	return
}
