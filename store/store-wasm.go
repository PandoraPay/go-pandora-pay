//go:build wasm
// +build wasm

package store

import (
	"errors"
	"pandora-pay/config/globals"
	store_db_bunt "pandora-pay/store/store-db/store-db-bunt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	store_db_js "pandora-pay/store/store-db/store-db-js"
	store_db_memory "pandora-pay/store/store-db/store-db-memory"
)

func createStoreNow(name string, storeType string) (*Store, error) {

	var db store_db_interface.StoreDBInterface
	var err error

	switch storeType {
	case "bunt-memory":
		db, err = store_db_bunt.CreateStoreDBBunt(name, true)
	case "js":
		db, err = store_db_js.CreateStoreDBJS(name)
	case "memory":
		db, err = store_db_memory.CreateStoreDBMemory(name)
	default:
		err = errors.New("Invalid --store-type argument: " + storeType)
	}

	if err != nil {
		return nil, err
	}

	return createStore(name, db)
}

func create_db() (err error) {

	var prefix = ""

	allowedStores := map[string]bool{"bunt-memory": true, "memory": true, "js": true}

	if StoreBlockchain, err = createStoreNow(prefix+"/blockchain", getStoreType(globals.Arguments["--store-chain-type"].(string), allowedStores)); err != nil {
		return
	}
	if StoreWallet, err = createStoreNow(prefix+"/wallet", getStoreType(globals.Arguments["--store-wallet-type"].(string), allowedStores)); err != nil {
		return
	}
	if StoreSettings, err = createStoreNow(prefix+"/settings", getStoreType(globals.Arguments["--store-wallet-type"].(string), allowedStores)); err != nil {
		return
	}
	if StoreMempool, err = createStoreNow(prefix+"/mempool", getStoreType(globals.Arguments["--store-wallet-type"].(string), allowedStores)); err != nil {
		return
	}

	return
}
