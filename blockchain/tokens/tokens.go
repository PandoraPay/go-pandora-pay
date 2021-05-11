package tokens

import (
	"errors"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store"
)

type Tokens struct {
	store.HashMap
}

func NewTokens(tx *bbolt.Tx) *Tokens {
	return &Tokens{
		HashMap: *store.CreateNewHashMap(tx, "Tokens", 20),
	}
}

func (tokens *Tokens) GetToken(key []byte) *token.Token {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	data := tokens.HashMap.Get(key)
	if data == nil {
		return nil
	}

	tok := new(token.Token)
	if err := tok.Deserialize(helpers.NewBufferReader(data)); err != nil {
		panic(err)
	}

	return tok
}

func (tokens *Tokens) CreateToken(key []byte, tok *token.Token) error {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	if err := tok.Validate(); err != nil {
		return err
	}
	if tokens.ExistsToken(key) {
		return errors.New("token already exists")
	}

	tokens.UpdateToken(key, tok)
	return nil
}

func (tokens *Tokens) UpdateToken(key []byte, tok *token.Token) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	tokens.HashMap.Update(key, tok.SerializeToBytes())
}

func (tokens *Tokens) ExistsToken(key []byte) bool {
	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	return tokens.HashMap.Exists(key)
}

func (tokens *Tokens) DeleteToken(key []byte) {
	tokens.HashMap.Delete(key)
}
