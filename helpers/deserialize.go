package helpers

import (
	"encoding/binary"
	"errors"
)

func DeserializeNumber(buf []byte) (uint64, []byte, error) {

	out, n := binary.Uvarint(buf)
	if n <= 0 {
		return 0, buf, errors.New("Deserializing Number exceeded")
	}

	return out, buf[n:], nil
}

func DeserializeBuffer(buf []byte, count int) ([]byte, []byte, error) {

	if count > len(buf) {
		return nil, buf, errors.New("Deserializing buffer exceeded")
	}

	return buf[:count], buf[count:], nil
}
