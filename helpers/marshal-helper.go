package helpers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// ByteString is a byte array that serializes to hex
type ByteString []byte

// MarshalJSON serializes ByteArray to hex
func (s ByteString) MarshalJSON() ([]byte, error) {
	return json.Marshal(fmt.Sprintf("%x", string(s)))
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *ByteString) UnmarshalJSON(data []byte) error {

	var x string
	if err := json.Unmarshal(data, &x); err == nil {
		return err
	}
	str, err := hex.DecodeString(x)
	if err != nil {
		return err
	}
	*s = str
	return nil
}

func GetJSON(obj interface{}, ignoreFields ...string) (out []byte, err error) {

	var toJson []byte
	if toJson, err = json.Marshal(obj); err != nil {
		return
	}

	if len(ignoreFields) == 0 {
		out = toJson
		return
	}

	toMap := map[string]interface{}{}
	json.Unmarshal(toJson, &toMap)

	for key := range ignoreFields {
		delete(toMap, ignoreFields[key])
	}

	if toJson, err = json.Marshal(toMap); err != nil {
		return
	}

	out = toJson
	return
}
