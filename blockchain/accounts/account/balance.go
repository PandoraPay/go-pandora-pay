package account

import (
	"errors"
	"pandora-pay/helpers"
)

type Balance struct {
	Amount uint64
	Token  []byte
}

func (balance *Balance) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUint64(balance.Amount)

	if len(balance.Token) == 0 {
		writer.WriteByte(0)
	} else {
		writer.WriteByte(1)
		writer.Write(balance.Token[:])
	}
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
