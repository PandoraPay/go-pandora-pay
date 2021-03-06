package tokens

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/tokens/token"
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

func (tokens *Tokens) GetToken(key [20]byte) *token.Token {

	data := tokens.HashMap.Get(key[:])
	if data == nil {
		return nil
	}

	tok := new(token.Token)
	tok.Deserialize(data)
	return tok
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

func (tokens *Tokens) CommitToStore() {
	tokens.HashMap.CommitToStore()
}
