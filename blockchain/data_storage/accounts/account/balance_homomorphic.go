package account

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
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
	return json.Marshal(fmt.Sprintf("%x", string(s.Amount.Serialize())))
}

// EncodeMsgpack serializes ElGamal into byteArray
func (s BalanceHomomorphic) EncodeMsgpack(enc *msgpack.Encoder) error {
	return enc.EncodeBytes(s.Amount.Serialize())
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *BalanceHomomorphic) UnmarshalJSON(data []byte) (err error) {
	var x string
	var str []byte

	if err = json.Unmarshal(data, &x); err != nil {
		return
	}
	if str, err = hex.DecodeString(x); err != nil {
		return
	}

	if s.Amount, err = new(crypto.ElGamal).Deserialize(str); err != nil {
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

func (balance *BalanceHomomorphic) AddBalance(encryptedAmount []byte) (err error) {
	panic("not implemented")
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
