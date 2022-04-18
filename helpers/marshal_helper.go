package helpers

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
)

// HexBytes is a byte array that serializes to hex
type Base64 []byte

// UnmarshalText for Gorilla Decoder
// see https://github.com/gorilla/schema/blob/8285576f31afd6804df356a38883f4fa05014373/decoder_test.go#L20
func (s *Base64) UnmarshalText(data []byte) (err error) {
	if *s, err = base64.StdEncoding.DecodeString(string(data)); err != nil {
		return
	}
	return
}

// EncodeMsgpack serializes ElGamal into byteArray
func (s *Base64) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(*s)
}

// DecodeMsgpack deserializes ByteArray to hex
func (s *Base64) DecodeMsgpack(dec *msgpack.Decoder) error {
	bytes, err := dec.DecodeBytes()
	if err != nil {
		return err
	}
	*s = bytes
	return nil
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
