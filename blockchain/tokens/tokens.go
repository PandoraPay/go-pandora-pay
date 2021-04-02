package tokens

import (
	"errors"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/store"
)

type Tokens struct {
	HashMap *store.HashMap
}

func NewTokens(tx *bbolt.Tx) (tokens *Tokens) {

	hashMap := store.CreateNewHashMap(tx, "Tokens", 20)

	tokens = new(Tokens)
	tokens.HashMap = hashMap
	return
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
	if err := tok.Deserialize(data); err != nil {
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

	tokens.HashMap.Update(key, tok.Serialize())
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

func (tokens *Tokens) Rollback() {
	tokens.HashMap.Rollback()
}

func (tokens *Tokens) Commit() {
	tokens.HashMap.Commit()
}

func (tokens *Tokens) WriteToStore() error {
	return tokens.HashMap.WriteToStore()
}

func (tokens *Tokens) WriteTransitionalChangesToStore(prefix string) error {
	return tokens.HashMap.WriteTransitionalChangesToStore(prefix)
}
func (tokens *Tokens) ReadTransitionalChangesFromStore(prefix string) error {
	return tokens.HashMap.ReadTransitionalChangesFromStore(prefix)
}
func (tokens *Tokens) DeleteTransitionalChangesFromStore(prefix string) error {
	return tokens.HashMap.DeleteTransitionalChangesFromStore(prefix)
}
