package helpers

type SerializableInterface interface {
	Serialize(writer *BufferWriter)
	SerializeToBytes() []byte
	Deserialize(reader *BufferReader) error
}
