package tokens

import (
	"errors"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Tokens struct {
	hash_map.HashMap `json:"-"`
}

func NewTokens(tx store_db_interface.StoreDBTransactionInterface) (tokens *Tokens) {
	tokens = &Tokens{
		HashMap: *hash_map.CreateNewHashMap(tx, "tokens", cryptography.PublicKeyHashHashSize),
	}
	tokens.HashMap.Deserialize = func(data []byte) (helpers.SerializableInterface, error) {
		var tok = &token.Token{}
		err := tok.Deserialize(helpers.NewBufferReader(data))
		return tok, err
	}
	return
}

func (tokens *Tokens) GetToken(key []byte) (tok *token.Token, err error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	data, err := tokens.HashMap.Get(string(key))
	if data == nil || err != nil {
		return
	}

	tok = data.(*token.Token)

	return
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

	tokens.UpdateToken(key, tok)
	return
}

func (tokens *Tokens) UpdateToken(key []byte, tok *token.Token) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	tokens.Update(string(key), tok)
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

func (hashMap *Tokens) WriteToStore() (err error) {

	if err = hashMap.HashMap.WriteToStore(); err != nil {
		return
	}

	return
}
