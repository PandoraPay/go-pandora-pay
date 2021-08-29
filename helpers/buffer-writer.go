package helpers

import (
	"encoding/binary"
)

type BufferWriter struct {
	array [][]byte
	len   int
	temp  []byte
}

func NewBufferWriter() *BufferWriter {
	temp := make([]byte, binary.MaxVarintLen64)
	return &BufferWriter{temp: temp}
}

func (writer *BufferWriter) Write(value []byte) {
	writer.array = append(writer.array, value)
	writer.len += len(value)
}

func (writer *BufferWriter) WriteString(string string) {
	value := []byte(string)
	writer.WriteUvarint(uint64(len(value)))
	writer.array = append(writer.array, value)
	writer.len += len(value)
}

func (writer *BufferWriter) WriteBool(value bool) {
	var value2 byte
	if value {
		value2 = 1
	}
	writer.array = append(writer.array, []byte{value2})
	writer.len += 1
}

func (writer *BufferWriter) WriteByte(value byte) {
	writer.array = append(writer.array, []byte{value})
	writer.len += 1
}

func (writer *BufferWriter) WriteUvarint(value uint64) {

	n := binary.PutUvarint(writer.temp, value)
	buf := make([]byte, n)
	copy(buf[:], writer.temp[:n])
	writer.array = append(writer.array, buf)
	writer.len += n

}

func (writer *BufferWriter) WriteToken(token []byte) {
	if len(token) == 0 {
		writer.WriteByte(0)
	} else {
		writer.WriteByte(1)
		writer.Write(token)
	}
}

func (writer *BufferWriter) Bytes() (out []byte) {
	out = make([]byte, writer.len)
	c := 0
	for i := 0; i < len(writer.array); i++ {
		copy(out[c:], writer.array[i])
		c += len(writer.array[i])
	}
	return
}

func (writer *BufferWriter) Length() int {
	return writer.len
}
