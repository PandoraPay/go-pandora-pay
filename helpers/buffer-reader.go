package helpers

import (
	"encoding/binary"
	"pandora-pay/cryptography"
)

type BufferReader struct {
	Buf      []byte
	Position int
}

func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{Buf: buf, Position: 0}
}

func (reader *BufferReader) ReadBool() bool {
	if len(reader.Buf) > 0 {
		if reader.Buf[0] > 1 {
			panic("buf[0] is invalid")
		}
		out := reader.Buf[0] == 1
		reader.Position += 1
		reader.Buf = reader.Buf[1:]
		return out
	}
	panic("Error reading bool")
}

func (reader *BufferReader) ReadByte() byte {
	if len(reader.Buf) > 0 {
		out := reader.Buf[0]
		reader.Position += 1
		reader.Buf = reader.Buf[1:]
		return out
	}
	panic("Error reading byte")
}

func (reader *BufferReader) ReadBytes(count int) []byte {
	if len(reader.Buf) >= count {
		out := reader.Buf[:count]
		reader.Position += count
		reader.Buf = reader.Buf[count:]
		return out
	}
	panic("Error reading bytes")
}

func (reader *BufferReader) ReadString() string {

	length := int(reader.ReadUvarint())
	reader.Position += length
	bytes := reader.ReadBytes(length)

	return string(bytes)
}

func (reader *BufferReader) ReadHash() []byte {
	if len(reader.Buf) >= cryptography.HashSize {
		out := reader.Buf[:cryptography.HashSize]
		reader.Position += cryptography.HashSize
		reader.Buf = reader.Buf[cryptography.HashSize:]
		return out
	}
	panic("Error reading hash")
}

func (reader *BufferReader) ReadToken() []byte {

	tokenType := reader.ReadByte()
	reader.Position += 1
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
	for i, b := range reader.Buf {
		if b < 0x80 {
			if i >= binary.MaxVarintLen64 || i == binary.MaxVarintLen64-1 && b > 1 {
				panic("Overflow")
			}
			reader.Position += i + 1
			reader.Buf = reader.Buf[i+1:]
			return x | uint64(b)<<s
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	panic("Error reading value")
}
