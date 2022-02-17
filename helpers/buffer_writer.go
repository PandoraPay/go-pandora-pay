package helpers

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"math"
	"pandora-pay/config/config_coins"
)

type BufferWriter struct {
	array [][]byte
	len   int
	temp  []byte
}

func (writer *BufferWriter) Write(value []byte) {
	writer.array = append(writer.array, value)
	writer.len += len(value)
}

func (writer *BufferWriter) WriteString(string string) {
	writer.WriteVariableBytes([]byte(string))
}

func (writer *BufferWriter) WriteVariableBytes(value []byte) {
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

func (writer *BufferWriter) WriteFloat64(value float64) {
	bits := math.Float64bits(value)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	writer.array = append(writer.array, bytes)
	writer.len += 8
}

func (writer *BufferWriter) WriteAsset(asset []byte) {
	if bytes.Equal(asset, config_coins.NATIVE_ASSET_FULL) {
		writer.WriteByte(0)
	} else {
		writer.WriteByte(1)
		writer.Write(asset)
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

func (writer *BufferWriter) Hex() string {
	data := writer.Bytes()
	return hex.EncodeToString(data)
}

func (writer *BufferWriter) Length() int {
	return writer.len
}

func NewBufferWriter() *BufferWriter {
	temp := make([]byte, binary.MaxVarintLen64)
	return &BufferWriter{temp: temp}
}
