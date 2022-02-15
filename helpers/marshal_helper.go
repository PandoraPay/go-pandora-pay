package helpers

import (
	"encoding/binary"
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/cryptography/bn256"
)

func ConvertBN256Array(array []*bn256.G1) [][]byte {
	out := make([][]byte, len(array))
	for i, it := range array {
		out[i] = it.EncodeCompressed()
	}
	return out
}

func ConvertToBN256Array(array [][]byte) ([]*bn256.G1, error) {
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

func GetMarshalledDataExcept(obj interface{}, ignoreFields ...string) ([]byte, error) {

	toJson, err := msgpack.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if len(ignoreFields) == 0 {
		return toJson, nil
	}

	toMap := map[string]interface{}{}
	if err = msgpack.Unmarshal(toJson, &toMap); err != nil {
		return nil, err
	}

	for key := range ignoreFields {
		delete(toMap, ignoreFields[key])
	}

	return msgpack.Marshal(toMap)
}

func GetJSONDataExcept(obj interface{}, ignoreFields ...string) ([]byte, error) {

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

	return json.Marshal(toMap)
}

func BytesLengthSerialized(value uint64) int {
	buf := make([]byte, binary.MaxVarintLen64)
	return binary.PutUvarint(buf, value)
}

func init() {
}
