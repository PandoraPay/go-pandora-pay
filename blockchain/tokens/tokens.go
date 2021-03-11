package tokens

import (
	"bytes"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store"
)

type Tokens struct {
	HashMap *store.HashMap
}

func NewTokens(tx *bbolt.Tx) (tokens *Tokens) {

	if tx == nil {
		panic("DB Transaction is not set")
	}

	hashMap := store.CreateNewHashMap(tx, "Tokens", 20)

	tokens = new(Tokens)
	tokens.HashMap = hashMap
	return
}

func (tokens *Tokens) GetAnyToken(key []byte) *token.Token {
	if bytes.Equal(key, config.NATIVE_TOKEN) {
		return tokens.GetToken(config.NATIVE_TOKEN_FULL)
	}
	return tokens.GetToken(*helpers.Byte20(key))
}

func (tokens *Tokens) GetToken(key [20]byte) *token.Token {

	data := tokens.HashMap.Get(key[:])
	if data == nil {
		return nil
	}

	tok := new(token.Token)
	tok.Deserialize(data)
	return tok
}

func (tokens *Tokens) CreateToken(key [20]byte, tok *token.Token) {
	tok.Validate()
	if tokens.ExistsToken(key) {
		panic("token already exists")
	}
	tokens.UpdateToken(key, tok)
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

func (tokens *Tokens) Rollback() {
	tokens.HashMap.Rollback()
}

func (tokens *Tokens) Commit() {
	tokens.HashMap.Commit()
}

func (tokens *Tokens) WriteToStore() {
	tokens.HashMap.WriteToStore()
}

func (tokens *Tokens) WriteTransitionalChangesToStore(prefix string) {
	tokens.HashMap.WriteTransitionalChangesToStore(prefix)
}

func (tokens *Tokens) ReadTransitionalChangesFromStore(prefix string) {
	tokens.HashMap.ReadTransitionalChangesFromStore(prefix)
}
