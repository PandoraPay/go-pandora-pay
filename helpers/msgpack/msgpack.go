package msgpack

import (
	m "github.com/vmihailenco/msgpack/v5"
)

func Marshal(v any) ([]byte, error) {
	return m.Marshal(v)
	//b := &bytes.Buffer{}
	//err := codec.NewEncoder(b, &codec.MsgpackHandle{}).Encode(v)
	//if err != nil {
	//	return nil, err
	//}
	//return b.Bytes(), nil
}

func Unmarshal(data []byte, v any) error {
	return m.Unmarshal(data, v)
	//return codec.NewDecoderBytes(data, &codec.MsgpackHandle{}).Decode(v)
}
