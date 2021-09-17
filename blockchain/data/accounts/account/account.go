package account

import (
	"errors"
	"math/big"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type Account struct {
	helpers.SerializableInterface `json:"-"`
	PublicKey                     []byte              `json:"-"` //hashmap key
	Token                         []byte              `json:"-"` //collection token
	Version                       uint64              `json:"version"`
	Balance                       *BalanceHomomorphic `json:"balances"`
}

func (account *Account) Validate() error {
	if account.Version != 0 {
		return errors.New("Version is invalid")
	}
	return nil
}

func (account *Account) AddBalanceUint(amount uint64) (err error) {
	account.Balance.Amount = account.Balance.Amount.Plus(new(big.Int).SetUint64(amount))
	return
}

func (account *Account) AddBalance(encryptedAmount []byte) (err error) {
	panic("not implemented")
}

func (account *Account) GetBalance() (result *crypto.ElGamal) {
	return account.Balance.Amount
}

func (account *Account) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(account.Version)
	account.Balance.Serialize(w)
}

func (account *Account) SerializeToBytes() []byte {
	w := helpers.NewBufferWriter()
	account.Serialize(w)
	return w.Bytes()
}

func (account *Account) Deserialize(r *helpers.BufferReader) (err error) {

	if account.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if err = account.Balance.Deserialize(r); err != nil {
		return
	}

	return
}

func NewAccount(publicKey []byte, token []byte) *Account {

	var acckey crypto.Point
	if err := acckey.DecodeCompressed(publicKey); err != nil {
		panic(err)
	}
	acc := &Account{
		PublicKey: publicKey,
		Token:     token,
		Balance:   &BalanceHomomorphic{crypto.ConstructElGamal(acckey.G1(), crypto.ElGamal_BASE_G)},
	}

	return acc
}
