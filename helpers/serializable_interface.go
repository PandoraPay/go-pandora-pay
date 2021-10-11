package helpers

type SerializableInterface interface {
	Serialize(w *BufferWriter)
	SerializeToBytes() []byte
	Deserialize(r *BufferReader) error
}
