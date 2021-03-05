package account

import (
	"pandora-pay/helpers"
)

type Balance struct {
	Amount uint64
	Token  []byte
}

func (balance *Balance) Serialize(writer *helpers.BufferWriter) {

	writer.WriteUvarint(balance.Amount)
	writer.WriteToken(balance.Token)

}

func (balance *Balance) Deserialize(reader *helpers.BufferReader) {
	balance.Amount = reader.ReadUvarint()
	balance.Token = reader.ReadToken()
	return
}
