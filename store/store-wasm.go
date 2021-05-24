// +build wasm

package store

import (
	"pandora-pay/config"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	store_db_wasm "pandora-pay/store/store-db/store-db-wasm"
)

func createStoreNow(name string) (store *Store, err error) {
	var db store_db_interface.StoreDBInterface
	if db, err = store_db_wasm.CreateStoreDBWASM(name); err != nil {
		return
	}

	store, err = createStore(name, db)
	return
}

func create_db() (err error) {

	var prefix = config.GetNetworkName()

	StoreBlockchain = &Store{Name: prefix + "/blockchain"}
	StoreWallet = &Store{Name: prefix + "/wallet"}
	StoreSettings = &Store{Name: prefix + "/settings"}
	StoreMempool = &Store{Name: prefix + "/mempool"}

	return
}
