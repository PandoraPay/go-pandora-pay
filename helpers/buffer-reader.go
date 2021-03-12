package helpers

import (
	"encoding/binary"
	"pandora-pay/cryptography"
)

type BufferReader struct {
	buf []byte
}

func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{buf: buf}
}

func (reader *BufferReader) ReadBool() bool {
	if len(reader.buf) > 0 {
		if reader.buf[0] > 1 {
			panic("buf[0] is invalid")
		}
		out := reader.buf[0] == 1
		reader.buf = reader.buf[1:]
		return out
	}
	panic("Error reading bool")
}

func (reader *BufferReader) ReadByte() byte {
	if len(reader.buf) > 0 {
		out := reader.buf[0]
		reader.buf = reader.buf[1:]
		return out
	}
	panic("Error reading byte")
}

func (reader *BufferReader) ReadBytes(count int) []byte {
	if len(reader.buf) >= count {
		out := reader.buf[:count]
		reader.buf = reader.buf[count:]
		return out
	}
	panic("Error reading bytes")
}

func (reader *BufferReader) ReadString() string {

	length := reader.ReadUvarint()
	bytes := reader.ReadBytes(int(length))

	return string(bytes)
}

func (reader *BufferReader) ReadHash() []byte {
	if len(reader.buf) >= cryptography.HashSize {
		out := reader.buf[:cryptography.HashSize]
		reader.buf = reader.buf[cryptography.HashSize:]
		return out
	}
	panic("Error reading hash")
}

func (reader *BufferReader) ReadToken() []byte {

	tokenType := reader.ReadByte()

	if tokenType == 0 {
		return []byte{}
	} else if tokenType == 1 {
		return reader.ReadBytes(20)
	}
	panic("invalid token type")
}

func (reader *BufferReader) ReadUvarint() uint64 {
	var x uint64
	var s uint
	for i, b := range reader.buf {
		if b < 0x80 {
			if i >= binary.MaxVarintLen64 || i == binary.MaxVarintLen64-1 && b > 1 {
				panic("Overflow")
			}
			reader.buf = reader.buf[i+1:]
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	panic("Error reading value")
}
