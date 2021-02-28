package helpers

import (
	"encoding/binary"
	"errors"
)

type BufferReader struct {
	buf []byte
}

func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{buf: buf}
}

func (reader *BufferReader) ReadBool() (out bool, err error) {
	if len(reader.buf) > 0 {
		out = reader.buf[0] == 1
		reader.buf = reader.buf[1:]
		return
	}
	err = errors.New("Error reading bool")
	return
}

func (reader *BufferReader) ReadByte() (out byte, err error) {
	if len(reader.buf) > 0 {
		out = reader.buf[0]
		reader.buf = reader.buf[1:]
		return
	}
	err = errors.New("Error reading byte")
	return
}

func (reader *BufferReader) ReadBytes(count int) (out []byte, err error) {
	if len(reader.buf) >= count {
		out = reader.buf[:count]
		reader.buf = reader.buf[count:]
		return
	}
	err = errors.New("Error reading bytes")
	return
}

func (reader *BufferReader) ReadHash() (out Hash, err error) {
	if len(reader.buf) > HashSize {
		out = *ConvertHash(reader.buf[:HashSize])
		reader.buf = reader.buf[HashSize:]
		return
	}
	err = errors.New("Error reading hash")
	return
}

func (reader *BufferReader) ReadUvarint() (uint64, error) {
	var x uint64
	var s uint
	for i, b := range reader.buf {
		if b < 0x80 {
			if i >= binary.MaxVarintLen64 || i == binary.MaxVarintLen64-1 && b > 1 {
				return 0, errors.New("Overflow")
			}
			reader.buf = reader.buf[i+1:]
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, errors.New("Error reading value")
}
