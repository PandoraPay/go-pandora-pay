package account

import (
	"encoding/json"
	"github.com/vmihailenco/msgpack/v5"
	"math/big"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type BalanceHomomorphic struct {
	helpers.SerializableInterface `json:"-" msgpack:"-"`
	Amount                        *crypto.ElGamal `json:"amount" msgpack:"amount"`
}

// MarshalJSON serializes ElGamal into byteArray
func (s BalanceHomomorphic) MarshalJSON() ([]byte, error) {
	serialized := s.Amount.Serialize()
	return json.Marshal(serialized)
}

// EncodeMsgpack serializes ElGamal into byteArray
func (s BalanceHomomorphic) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(s.Amount.Serialize())
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *BalanceHomomorphic) UnmarshalJSON(data []byte) (err error) {

	var serialized []byte
	if err = json.Unmarshal(data, &serialized); err != nil {
		return
	}
	if s.Amount, err = new(crypto.ElGamal).Deserialize(serialized); err != nil {
		return
	}
	return
}

// DecodeMsgpack deserializes ByteArray to hex
func (s *BalanceHomomorphic) DecodeMsgpack(dec *msgpack.Decoder) error {
	bytes, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	if s.Amount, err = new(crypto.ElGamal).Deserialize(bytes); err != nil {
		return err
	}
	return nil
}

func (balance *BalanceHomomorphic) AddBalanceUint(amount uint64) {
	balance.Amount = balance.Amount.Plus(new(big.Int).SetUint64(amount))
}

func (balance *BalanceHomomorphic) Serialize(w *helpers.BufferWriter) {
	w.Write(balance.Amount.Serialize())
}

func (balance *BalanceHomomorphic) Deserialize(r *helpers.BufferReader) (err error) {

	var amount []byte
	if amount, err = r.ReadBytes(66); err != nil {
		return
	}
	if balance.Amount, err = new(crypto.ElGamal).Deserialize(amount); err != nil {
		return
	}

	return
}
