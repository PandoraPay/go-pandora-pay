package helpers

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"pandora-pay/cryptography/bn256"
)

// HexBytes is a byte array that serializes to hex
type HexBytes []byte

// MarshalJSON serializes ByteArray to hex
func (s HexBytes) MarshalJSON() ([]byte, error) {
	dst := make([]byte, len(s)*2+2)
	hex.Encode(dst[1:], s)
	dst[0] = 34          // "
	dst[len(dst)-1] = 34 // "
	return dst, nil
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *HexBytes) UnmarshalJSON(data []byte) (err error) {

	str := make([]byte, len(data)/2-1)

	if _, err = hex.Decode(str, data[1:len(data)-1]); err != nil {
		return
	}
	*s = str
	return
}

func ConvertHexBytesArraysToBytesArray(data []HexBytes) [][]byte {
	out := make([][]byte, len(data))
	for i := range data {
		out[i] = data[i]
	}
	return out
}

func ConvertBN256Array(array []*bn256.G1) []HexBytes {
	out := make([]HexBytes, len(array))
	for i, it := range array {
		out[i] = it.EncodeCompressed()
	}
	return out
}

func ConvertToBN256Array(array []HexBytes) ([]*bn256.G1, error) {
	out := make([]*bn256.G1, len(array))
	for i := range array {

		p := new(bn256.G1)
		if err := p.DecodeCompressed(array[i]); err != nil {
			return nil, err
		}
		out[i] = p
	}
	return out, nil
}

func GetJSON(obj interface{}, ignoreFields ...string) ([]byte, error) {

	toJson, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if len(ignoreFields) == 0 {
		return toJson, nil
	}

	toMap := map[string]interface{}{}
	if err = json.Unmarshal(toJson, &toMap); err != nil {
		return nil, err
	}

	for key := range ignoreFields {
		delete(toMap, ignoreFields[key])
	}

	if toJson, err = json.Marshal(toMap); err != nil {
		return nil, err
	}

	return toJson, nil
}

func BytesLengthSerialized(value uint64) int {
	buf := make([]byte, binary.MaxVarintLen64)
	return binary.PutUvarint(buf, value)
}
