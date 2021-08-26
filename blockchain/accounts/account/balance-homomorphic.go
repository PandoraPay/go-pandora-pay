package account

import (
	"pandora-pay/cryptography/cryptolib"
	"pandora-pay/helpers"
)

type BalanceHomomorphic struct {
	Amount *cryptolib.ElGamal `json:"amount"`
	Token  helpers.HexBytes   `json:"token"` //20
}

func (balance *BalanceHomomorphic) Serialize(writer *helpers.BufferWriter) {
	writer.Write(balance.Amount.Serialize())
	writer.WriteToken(balance.Token)
}

func (balance *BalanceHomomorphic) Deserialize(reader *helpers.BufferReader) (err error) {

	var amount []byte
	if amount, err = reader.ReadBytes(66); err != nil {
		return
	}
	balance.Amount = new(cryptolib.ElGamal).Deserialize(amount)
	if balance.Token, err = reader.ReadToken(); err != nil {
		return
	}

	return
}
