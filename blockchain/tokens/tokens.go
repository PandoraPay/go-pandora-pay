package tokens

import (
	"errors"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Tokens struct {
	hash_map.HashMap `json:"-"`
}

func NewTokens(tx store_db_interface.StoreDBTransactionInterface) *Tokens {
	return &Tokens{
		HashMap: *hash_map.CreateNewHashMap(tx, "Tokens", 20),
	}
}

func (tokens *Tokens) GetToken(key []byte) (tok *token.Token, err error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	data := tokens.HashMap.Get(key)
	if data == nil {
		return
	}

	tok = new(token.Token)
	if err = tok.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return
	}

	return
}

func (tokens *Tokens) CreateToken(key []byte, tok *token.Token) (err error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	if err = tok.Validate(); err != nil {
		return
	}
	if tokens.ExistsToken(key) {
		return errors.New("token already exists")
	}

	tokens.UpdateToken(key, tok)
	return
}

func (tokens *Tokens) UpdateToken(key []byte, tok *token.Token) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	tokens.Update(key, tok.SerializeToBytes())
}

func (tokens *Tokens) ExistsToken(key []byte) bool {
	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	return tokens.Exists(key)
}

func (tokens *Tokens) DeleteToken(key []byte) {
	tokens.Delete(key)
}
