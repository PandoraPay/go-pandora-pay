package helpers

type SerializableInterface interface {
	Validate() error
	Serialize(w *BufferWriter)
	SerializeToBytes() []byte
	Deserialize(r *BufferReader) error
}
