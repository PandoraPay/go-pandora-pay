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

func (reader *BufferReader) ReadBool() (bool, error) {
	if len(reader.buf) > 0 {
		out := reader.buf[0] == 1
		reader.buf = reader.buf[1:]
		return out, nil
	}
	return false, errors.New("Error reading bool")
}

func (reader *BufferReader) ReadByte() (byte, error) {
	if len(reader.buf) > 0 {
		out := reader.buf[0]
		reader.buf = reader.buf[1:]
		return out, nil
	}
	return 0, errors.New("Error reading byte")
}

func (reader *BufferReader) ReadBytes(count int) ([]byte, error) {
	if len(reader.buf) >= count {
		out := reader.buf[:count]
		reader.buf = reader.buf[count:]
		return out, nil
	}
	return nil, errors.New("Error reading bytes")
}

func (reader *BufferReader) ReadString() (str string, err error) {

	var length uint64
	if length, err = reader.ReadUvarint(); err != nil {
		return
	}

	var bytes []byte
	if bytes, err = reader.ReadBytes(int(length)); err != nil {
		return
	}
	str = string(bytes)

	return
}

func (reader *BufferReader) ReadHash() (Hash, error) {
	if len(reader.buf) >= HashSize {
		out := *ConvertHash(reader.buf[:HashSize])
		reader.buf = reader.buf[HashSize:]
		return out, nil
	}
	return Hash{}, errors.New("Error reading hash")
}

func (reader *BufferReader) Read33() ([33]byte, error) {
	if len(reader.buf) >= 33 {
		out := *Byte33(reader.buf[:33])
		reader.buf = reader.buf[33:]
		return out, nil
	}
	return [33]byte{}, errors.New("Error reading 33byte")
}
func (reader *BufferReader) Read20() ([20]byte, error) {
	if len(reader.buf) >= 20 {
		out := *Byte20(reader.buf[:20])
		reader.buf = reader.buf[20:]
		return out, nil
	}
	return [20]byte{}, errors.New("Error reading 20byte")
}
func (reader *BufferReader) Read65() ([65]byte, error) {
	if len(reader.buf) >= 65 {
		out := *Byte65(reader.buf[:65])
		reader.buf = reader.buf[65:]
		return out, nil
	}
	return [65]byte{}, errors.New("Error reading 65byte ")
}

func (reader *BufferReader) ReadToken() (out []byte, err error) {

	var tokenType byte
	if tokenType, err = reader.ReadByte(); err != nil {
		return
	}

	if tokenType == 0 {
		out = []byte{}
	} else if tokenType == 1 {
		out, err = reader.ReadBytes(20)
	} else {
		err = errors.New("invalid token type")
		return
	}

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
