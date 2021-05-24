package store_db_js_indexdb

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBJSIndexDBTransaction struct {
	store_db_interface.StoreDBTransactionInterface
}

func (tx *StoreDBJSIndexDBTransaction) Put(key []byte, value []byte) error {
	return nil
}

func (tx *StoreDBJSIndexDBTransaction) Get(key []byte) []byte {
	return nil
}

func (tx *StoreDBJSIndexDBTransaction) Delete(key []byte) error {
	return nil
}
