package helpers

type SerializableInterface interface {
	Validate() error
	Serialize(w *BufferWriter)
	Deserialize(r *BufferReader) error
}

func SerializeToBytes(self SerializableInterface) []byte {
	w := NewBufferWriter()
	self.Serialize(w)
	return w.Bytes()
}
