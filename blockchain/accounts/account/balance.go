package account

import (
	"pandora-pay/helpers"
)

type Balance struct {
	Amount uint64
	Token  []byte
}

func (balance *Balance) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUint64(balance.Amount)
	writer.WriteToken(balance.Token)

}

func (balance *Balance) Deserialize(reader *helpers.BufferReader) (err error) {

	if balance.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}

	if balance.Token, err = reader.ReadToken(); err != nil {
		return
	}

	return
}
