package helpers

import (
	"encoding/binary"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
)

type BufferReader struct {
	Buf      []byte
	Position int
}

func NewBufferReader(buf []byte) *BufferReader {
	return &BufferReader{Buf: buf, Position: 0}
}

func (reader *BufferReader) ReadBool() (bool, error) {
	if len(reader.Buf) > 0 {
		if reader.Buf[reader.Position] > 1 {
			return false, errors.New("buf[0] is invalid")
		}
		out := reader.Buf[reader.Position] == 1
		reader.Position += 1
		return out, nil
	}
	return false, errors.New("Error reading bool")
}

func (reader *BufferReader) ReadByte() (byte, error) {
	if len(reader.Buf) > 0 {
		out := reader.Buf[reader.Position]
		reader.Position += 1
		return out, nil
	}
	return 0, errors.New("Error reading byte")
}

func (reader *BufferReader) ReadBytes(count int) ([]byte, error) {
	if len(reader.Buf) >= count {
		out := reader.Buf[reader.Position : reader.Position+count]
		reader.Position += count
		return out, nil
	}
	return nil, errors.New("Error reading bytes")
}

func (reader *BufferReader) ReadString() (string, error) {
	length, err := reader.ReadUvarint()
	if err != nil {
		return "", err
	}
	var bytes []byte
	if bytes, err = reader.ReadBytes(int(length)); err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (reader *BufferReader) ReadHash() ([]byte, error) {
	if len(reader.Buf) >= cryptography.HashSize {
		out := reader.Buf[reader.Position : reader.Position+cryptography.HashSize]
		reader.Position += cryptography.HashSize
		return out, nil
	}
	return nil, errors.New("Error reading hash")
}

func (reader *BufferReader) ReadToken() ([]byte, error) {
	tokenType, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}

	if tokenType == 0 {
		return []byte{}, nil
	} else if tokenType == 1 {
		return reader.ReadBytes(config.TOKEN_LENGTH)
	}
	return nil, errors.New("invalid token type")
}

func (reader *BufferReader) ReadUvarint() (uint64, error) {
	var x uint64
	var s uint

	var c byte
	for i := reader.Position; i < len(reader.Buf); i++ {
		b := reader.Buf[i]
		if b < 0x80 {
			if c >= binary.MaxVarintLen64 || c == binary.MaxVarintLen64-1 && b > 1 {
				return 0, errors.New("Overflow")
			}
			reader.Position = i + 1
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
		c += 1
	}
	return 0, errors.New("Error reading value")
}
