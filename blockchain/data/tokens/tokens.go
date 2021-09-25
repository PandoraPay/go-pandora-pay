package tokens

import (
	"errors"
	"pandora-pay/blockchain/data/tokens/token"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Tokens struct {
	hash_map.HashMap `json:"-"`
}

func (tokens *Tokens) GetToken(key []byte) (*token.Token, error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	data, err := tokens.HashMap.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*token.Token), nil
}

func (tokens *Tokens) CreateToken(key []byte, tok *token.Token) (err error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	if err = tok.Validate(); err != nil {
		return
	}

	var exists bool
	if exists, err = tokens.ExistsToken(key); err != nil {
		return
	}
	if exists {
		return errors.New("token already exists")
	}

	if err = tokens.UpdateToken(key, tok); err != nil {
		return
	}
	return
}

func (tokens *Tokens) UpdateToken(key []byte, tok *token.Token) error {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	return tokens.Update(string(key), tok)
}

func (tokens *Tokens) ExistsToken(key []byte) (bool, error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	return tokens.Exists(string(key))
}

func (tokens *Tokens) DeleteToken(key []byte) {
	tokens.Delete(string(key))
}

func NewTokens(tx store_db_interface.StoreDBTransactionInterface) (tokens *Tokens) {

	hashMap := hash_map.CreateNewHashMap(tx, "tokens", config.TOKEN_LENGTH, true)

	tokens = &Tokens{
		HashMap: *hashMap,
	}
	tokens.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var tok = &token.Token{}
		if err := tok.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return tok, nil
	}
	return
}
