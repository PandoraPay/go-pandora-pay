package tokens

import (
	"errors"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/store"
)

type Tokens struct {
	HashMap *store.HashMap
}

func NewTokens(tx *bbolt.Tx) (tokens *Tokens, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	var hashMap *store.HashMap
	if hashMap, err = store.CreateNewHashMap(tx, "Tokens", 20); err != nil {
		return
	}

	tokens = new(Tokens)
	tokens.HashMap = hashMap
	return
}

func (tokens *Tokens) GetToken(key [20]byte) (tok *token.Token, err error) {

	data := tokens.HashMap.Get(key[:])
	if data == nil {
		return
	}

	tok = new(token.Token)
	err = tok.Deserialize(data)
	return
}

func (tokens *Tokens) UpdateToken(key [20]byte, tok *token.Token) {
	tokens.HashMap.Update(key[:], tok.Serialize())
}

func (tokens *Tokens) ExistsToken(key [20]byte) bool {
	return tokens.HashMap.Exists(key[:])
}

func (tokens *Tokens) DeleteAccount(key [20]byte) {
	tokens.HashMap.Delete(key[:])
}

func (tokens *Tokens) Commit() error {
	return tokens.HashMap.Commit()
}
