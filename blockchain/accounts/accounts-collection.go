package accounts

import (
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
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

		var err error
		if accs, err = NewAccounts(collection.tx, token); err != nil {
			return nil, err
		}
		collection.accsMap[string(token)] = accs
	}
	return accs, nil
}

func (collection *AccountsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
	for _, accs := range collection.accsMap {
		accs.Tx = tx
	}
}

func (collection *AccountsCollection) Rollback() {
	for _, accs := range collection.accsMap {
		accs.Rollback()
	}
}

func (collection *AccountsCollection) CommitChanges() {
	for _, accs := range collection.accsMap {
		accs.CommitChanges()
	}
}

func (collection *AccountsCollection) WriteToStore() (err error) {
	for _, accs := range collection.accsMap {
		if err = accs.WriteToStore(); err != nil {
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

func NewAccountsCollection(tx store_db_interface.StoreDBTransactionInterface) *AccountsCollection {
	return &AccountsCollection{
		tx,
		make(map[string]*Accounts),
	}
}
