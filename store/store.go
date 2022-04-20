package store

import (
	"pandora-pay/store/store_db/store_db_interface"
)

type Store struct {
	Name   string
	Opened bool
	DB     store_db_interface.StoreDBInterface
}

var StoreBlockchain, StoreWallet, StoreSettings, StoreMempool *Store

func (store *Store) close() error {
	return store.DB.Close()
}

func createStore(name string, db store_db_interface.StoreDBInterface) (*Store, error) {

	store := &Store{
		Name:   name,
		Opened: false,
		DB:     db,
	}

	store.Opened = true

	return store, nil
}

func InitDB() (err error) {
	return create_db()
}

func DBClose() (err error) {
	if err = StoreBlockchain.close(); err != nil {
		return
	}
	if err = StoreWallet.close(); err != nil {
		return
	}
	if err = StoreSettings.close(); err != nil {
		return
	}
	if err = StoreMempool.close(); err != nil {
		return
	}
	return
}

func getStoreType(value string, allowed map[string]bool) string {

	if allowed[value] {
		return value
	}

	return ""
}
