package store_db_js_indexdb

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBJSIndexDB struct {
	store_db_interface.StoreDBInterface
	Name []byte
}

func (store *StoreDBJSIndexDB) Close() error {
	return nil
}

func (store *StoreDBJSIndexDB) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBJSIndexDBTransaction{}
	return callback(tx)
}

func (store *StoreDBJSIndexDB) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBJSIndexDBTransaction{}
	return callback(tx)
}

func CreateStoreDBJSIndexDB(name string) (store *StoreDBJSIndexDB, err error) {

	store = &StoreDBJSIndexDB{
		Name: []byte(name),
	}

	return
}
