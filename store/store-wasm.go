// +build wasm

package store

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	store_db_wasm "pandora-pay/store/store-db/store-db-wasm"
)

func createStoreNow(name string) (store *Store, err error) {
	var db store_db_interface.StoreDBInterface
	if db, err = store_db_wasm.CreateStoreDBJSIndexDB(name); err != nil {
		return
	}

	store, err = createStore(name, db)
	return
}

func create_db() (err error) {

	var prefix = ""

	if StoreBlockchain, err = createStoreNow(prefix + "/blockchain"); err != nil {
		return
	}
	if StoreWallet, err = createStoreNow(prefix + "/wallet"); err != nil {
		return
	}
	if StoreSettings, err = createStoreNow(prefix + "/settings"); err != nil {
		return
	}
	if StoreMempool, err = createStoreNow(prefix + "/mempool"); err != nil {
		return
	}

	return
}
