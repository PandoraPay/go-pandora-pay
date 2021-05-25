// +build wasm

package store

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

func createStoreNow(name string, storeType string) (store *Store, err error) {
	var db store_db_interface.StoreDBInterface

	switch storeType {
	case "memory":
		db, err = store_db_bunt.CreateStoreDBBunt(name, true)
	case "indexdb":
		if db, err = store_db_wasm.CreateStoreDBJSIndexDB(name); err != nil {
			return
		}
	default:
		err = errors.New("Invalid --store-type argument")
	}

	if err != nil {
		return
	}

	store, err = createStore(name, db)
	return
}

func create_db() (err error) {

	var prefix = ""

	if StoreBlockchain, err = createStoreNow(prefix+"/blockchain", getStoreType(globals.Arguments["--store-chain-type"], false, false, true, true, "memory")); err != nil {
		return
	}
	if StoreWallet, err = createStoreNow(prefix+"/wallet", getStoreType(globals.Arguments["--store-wallet-type"], false, false, true, true, "indexdb")); err != nil {
		return
	}
	if StoreSettings, err = createStoreNow(prefix+"/settings", getStoreType(globals.Arguments["--store-wallet-type"], false, false, true, true, "indexdb")); err != nil {
		return
	}
	if StoreMempool, err = createStoreNow(prefix+"/mempool", getStoreType(globals.Arguments["--store-wallet-type"], false, false, true, true, "indexdb")); err != nil {
		return
	}

	return
}
