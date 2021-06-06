package tokens

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/tokens/token"
	token_info "pandora-pay/blockchain/tokens/token-info"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Tokens struct {
	hash_map.HashMap `json:"-"`
}

func NewTokens(tx store_db_interface.StoreDBTransactionInterface) *Tokens {
	return &Tokens{
		HashMap: *hash_map.CreateNewHashMap(tx, "Tokens", cryptography.PublicKeyHashHashSize),
	}
}

func (tokens *Tokens) GetToken(key []byte) (tok *token.Token, err error) {

	if len(key) == 0 {
		key = config.NATIVE_TOKEN_FULL
	}

	data := tokens.HashMap.Get(string(key))
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

	tokens.Update(string(key), tok.SerializeToBytes())
}

func (tokens *Tokens) ExistsToken(key []byte) bool {
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

	if config.SEED_WALLET_NODES_INFO {
		for k, v := range hashMap.Committed {

			if v.Status == "del" {
				err = hashMap.Tx.DeleteForcefully("tokenInfo_byHash" + k)
			} else if v.Status == "update" {
				tok := &token.Token{}
				if err = json.Unmarshal(v.Data, tok); err != nil {
					return
				}

				tokInfo := &token_info.TokenInfo{
					Hash:             []byte(k),
					Name:             tok.Name,
					Ticker:           tok.Ticker,
					DecimalSeparator: tok.DecimalSeparator,
					Description:      tok.Description,
				}
				var data []byte
				data, err = json.Marshal(tokInfo)

				err = hashMap.Tx.Put("tokenInfo_byHash"+k, data)
			}

			if err != nil {
				return
			}

		}
	}

	return
}
