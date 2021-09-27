package accounts

import (
	"errors"
	"pandora-pay/blockchain/data/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

type Accounts struct {
	hash_map.HashMap `json:"-"`
	Token            []byte
}

func (accounts *Accounts) CreateAccount(publicKey []byte) (*account.Account, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	acc := account.NewAccount(publicKey, accounts.Token)
	if err := accounts.Update(string(publicKey), acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (accounts *Accounts) GetAccount(key []byte) (*account.Account, error) {
	data, err := accounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*account.Account), nil
}

func (accounts *Accounts) GetRandomAccount() (*account.Account, error) {
	data, err := accounts.GetRandom()
	if err != nil {
		return nil, err
	}
	return data.(*account.Account), nil
}

func (accounts *Accounts) saveTokensCount(key []byte, sign bool) (uint64, error) {

	var count uint64
	var err error

	data := accounts.Tx.Get("accounts:tokensCount:" + string(key))
	if data != nil {
		if count, err = helpers.NewBufferReader(data).ReadUvarint(); err != nil {
			return 0, err
		}
	}

	var countOriginal uint64
	if sign {
		countOriginal = count
		count += 1
	} else {
		count -= 1
		countOriginal = count
	}

	if count > 0 {
		w := helpers.NewBufferWriter()
		w.WriteUvarint(count)
		err = accounts.Tx.Put("accounts:tokensCount:"+string(key), w.Bytes())
	} else {
		err = accounts.Tx.Delete("accounts:tokensCount:" + string(key))
	}

	if err != nil {
		return 0, err
	}

	return countOriginal, nil
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface, Token []byte) (accounts *Accounts) {

	hashmap := hash_map.CreateNewHashMap(tx, "accounts", cryptography.PublicKeySize, true)

	accounts = &Accounts{
		HashMap: *hashmap,
		Token:   Token,
	}

	accounts.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var acc = account.NewAccount(key, accounts.Token)
		if err := acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return acc, nil
	}

	accounts.HashMap.StoredEvent = func(key []byte, element *hash_map.CommittedMapElement) (err error) {

		if !tx.IsWritable() {
			return
		}

		var count uint64
		if count, err = accounts.saveTokensCount(key, true); err != nil {
			return
		}

		return tx.Put("accounts:tokenByIndex:"+string(key)+":"+strconv.FormatUint(count, 10), element.Element.(*account.Account).Token)
	}

	accounts.HashMap.DeletedEvent = func(key []byte) (err error) {

		if !tx.IsWritable() {
			return
		}

		var count uint64
		if count, err = accounts.saveTokensCount(key, false); err != nil {
			return
		}

		return tx.Delete("accounts:tokenByIndex:" + string(key) + ":" + strconv.FormatUint(count, 10))
	}

	return
}
