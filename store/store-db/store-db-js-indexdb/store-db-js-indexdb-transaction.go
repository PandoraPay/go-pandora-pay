package store_db_js_indexdb

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBJSIndexDBTransaction struct {
	store_db_interface.StoreDBTransactionInterface
}

func (tx *StoreDBJSIndexDBTransaction) Put(key string, value []byte) error {
	return nil
}

func (tx *StoreDBJSIndexDBTransaction) Get(key string) []byte {
	return nil
}

func (tx *StoreDBJSIndexDBTransaction) Delete(key string) error {
	return nil
}
