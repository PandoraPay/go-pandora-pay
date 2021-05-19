package store

import (
	"os"
	"pandora-pay/config"
	"pandora-pay/context"
	store_db_bolt "pandora-pay/store/store-db/store-db-bolt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Store struct {
	Name   string
	Opened bool
	DB     store_db_interface.StoreDBInterface
}

var StoreBlockchain, StoreWallet, StoreSettings, StoreMempool *Store

func (store *Store) init() {
	var err error
	store.DB, err = store_db_bolt.CreateStoreDBBolt(store.Name)
	if err != nil {
		context.GUI.Fatal(err)
	}
	context.GUI.Log("Store Opened " + store.Name)
}

func (store *Store) close() (err error) {
	if err = store.DB.Close(); err != nil {
		return
	}
	context.GUI.Log("Store Closed " + store.Name)
	return
}

func DBInit() (err error) {

	var prefix = ""

	switch config.NETWORK_SELECTED {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix += "main"
	case config.TEST_NET_NETWORK_BYTE:
		prefix += "test"
	case config.DEV_NET_NETWORK_BYTE:
		prefix += "dev"
	default:
		panic("Network is unknown")
	}

	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	StoreBlockchain = &Store{Name: prefix + "/blockchain"}
	StoreWallet = &Store{Name: prefix + "/wallet"}
	StoreSettings = &Store{Name: prefix + "/settings"}
	StoreMempool = &Store{Name: prefix + "/mempool"}

	StoreBlockchain.init()
	StoreWallet.init()
	StoreSettings.init()
	StoreMempool.init()

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
