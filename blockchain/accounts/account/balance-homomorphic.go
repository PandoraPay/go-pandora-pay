package account

import (
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type BalanceHomomorphic struct {
	Amount *crypto.ElGamal  `json:"amount"`
	Token  helpers.HexBytes `json:"token"` //20
}

func (balance *BalanceHomomorphic) Serialize(w *helpers.BufferWriter) {
	w.Write(balance.Amount.Serialize())
	w.WriteToken(balance.Token)
}

func (balance *BalanceHomomorphic) Deserialize(r *helpers.BufferReader) (err error) {

	var amount []byte
	if amount, err = r.ReadBytes(66); err != nil {
		return
	}
	if balance.Amount, err = new(crypto.ElGamal).Deserialize(amount); err != nil {
		return
	}
	if balance.Token, err = r.ReadToken(); err != nil {
		return
	}

	return
}
