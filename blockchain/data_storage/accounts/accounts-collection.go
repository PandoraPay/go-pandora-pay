package accounts

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

type AccountsCollection struct {
	tx      store_db_interface.StoreDBTransactionInterface
	accsMap map[string]*Accounts
}

func (collection *AccountsCollection) GetAllMap() map[string]*Accounts {
	return collection.accsMap
}

func (collection *AccountsCollection) GetExistingMap(token []byte) (*Accounts, error) {
	if len(token) == 0 {
		token = config.NATIVE_TOKEN_FULL
	}
	if len(token) != cryptography.RipemdSize {
		return nil, errors.New("Token was not found")
	}
	return collection.accsMap[string(token)], nil
}

func (collection *AccountsCollection) GetMap(token []byte) (*Accounts, error) {

	if len(token) == 0 {
		token = config.NATIVE_TOKEN_FULL
	}

	if len(token) != cryptography.RipemdSize {
		return nil, errors.New("Token was not found")
	}

	accs := collection.accsMap[string(token)]
	if accs == nil {
		accs = NewAccounts(collection.tx, token)
		collection.accsMap[string(token)] = accs
	}
	return accs, nil
}

func (collection *AccountsCollection) GetAccountTokensCount(key []byte) (uint64, error) {

	var count uint64
	var err error

	data := collection.tx.Get("accounts:tokensCount:" + string(key))
	if data != nil {
		if count, err = helpers.NewBufferReader(data).ReadUvarint(); err != nil {
			return 0, err
		}
	}

	return count, nil
}

func (collection *AccountsCollection) GetAccountTokens(key []byte) ([][]byte, error) {

	count, err := collection.GetAccountTokensCount(key)
	if err != nil {
		return nil, err
	}

	out := make([][]byte, count)

	for i := uint64(0); i < count; i++ {
		token := collection.tx.Get("accounts:tokenByIndex:" + string(key) + ":" + strconv.FormatUint(i, 10))
		if token == nil {
			return nil, errors.New("Error reading token")
		}
		out[i] = token
	}

	return out, nil
}

func (collection *AccountsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
	for _, accs := range collection.accsMap {
		accs.SetTx(tx)
	}
}

func (collection *AccountsCollection) Rollback() {
	for _, accs := range collection.accsMap {
		accs.Rollback()
	}
}

func (collection *AccountsCollection) CloneCommitted() (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.CloneCommitted(); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) CommitChanges() (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.CommitChanges(); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) WriteTransitionalChangesToStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.WriteTransitionalChangesToStore(prefix); err != nil {
			return
		}
	}
	return
}

func (collection *AccountsCollection) ReadTransitionalChangesFromStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.ReadTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}
func (collection *AccountsCollection) DeleteTransitionalChangesFromStore(prefix string) (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.DeleteTransitionalChangesFromStore(prefix); err != nil {
			return
		}
	}
	return
}

func NewAccountsCollection(tx store_db_interface.StoreDBTransactionInterface) *AccountsCollection {
	return &AccountsCollection{
		tx,
		make(map[string]*Accounts),
	}
}
