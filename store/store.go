package store

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
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

func createStore(name string, db store_db_interface.StoreDBInterface) (store *Store, err error) {

	store = &Store{
		Name:   name,
		Opened: false,
		DB:     db,
	}

	store.Opened = true

	return
}

func InitDB() (err error) {
	if err = create_db(); err != nil {
		return
	}
	return
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
