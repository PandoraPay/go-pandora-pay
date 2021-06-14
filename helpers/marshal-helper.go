package helpers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// HexBytes is a byte array that serializes to hex
type HexBytes []byte

// MarshalJSON serializes ByteArray to hex
func (s HexBytes) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", string(s)))
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *HexBytes) UnmarshalJSON(data []byte) (err error) {

	var x string
	var str []byte

	if err = json.Unmarshal(data, &x); err != nil {
		return
	}
	if str, err = hex.DecodeString(x); err != nil {
		return
	}
	*s = str
	return
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
