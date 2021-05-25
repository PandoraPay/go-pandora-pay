// +build !wasm

package store

import (
	"errors"
	"pandora-pay/config/globals"
	store_db_bolt "pandora-pay/store/store-db/store-db-bolt"
	store_db_bunt "pandora-pay/store/store-db/store-db-bunt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

func createStoreNow(name string) (store *Store, err error) {
	var db store_db_interface.StoreDBInterface

	if globals.Arguments["--store-type"] == nil {
		globals.Arguments["--store-type"] = "bolt"
	}
	storeType := globals.Arguments["--store-type"].(string)

	switch storeType {
	case "bolt":
		db, err = store_db_bolt.CreateStoreDBBolt(name)
	case "bunt":
		db, err = store_db_bunt.CreateStoreDBBunt(name, false)
	case "memory":
		db, err = store_db_bunt.CreateStoreDBBunt(name, true)
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
