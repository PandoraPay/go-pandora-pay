package account

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/helpers"
)

type Balance struct {
	Amount uint64
	Token  []byte
}

func (balance *Balance) Serialize(serialized *bytes.Buffer, temp []byte) {

	n := binary.PutUvarint(temp, balance.Amount)
	serialized.Write(temp[:n])

	if len(balance.Token) == 0 {
		serialized.Write([]byte{0})
	} else {
		serialized.Write([]byte{1})
		serialized.Write(balance.Token[:])
	}

	serialized.Write(temp[:1])

}

func (balance *Balance) Deserialize(reader *helpers.BufferReader) (err error) {

	if balance.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}

	var tokenType byte
	if tokenType, err = reader.ReadByte(); err != nil {
		return
	}

	if tokenType == 0 {
		balance.Token = []byte{}
	} else if tokenType == 1 {
		if balance.Token, err = reader.ReadBytes(20); err != nil {
			return
		}
	} else {
		err = errors.New("invalid token type")
		return
	}

	return
}
