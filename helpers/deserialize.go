package helpers

import (
	"encoding/binary"
	"errors"
	"pandora-pay/crypto"
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

func DeserializeHash(buf []byte, count int) (crypto.Hash, []byte, error) {

	var out = crypto.Hash{}

	if count > len(buf) {
		return out, buf, errors.New("Deserializing buffer exceeded")
	}

	return *crypto.ConvertHash(buf[count:]), buf[count:], nil
}
