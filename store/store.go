package store

import (
	bolt "go.etcd.io/bbolt"
	"os"
	"pandora-pay/config"
	"pandora-pay/gui"
)

type Store struct {
	Name   string
	Opened bool
	DB     *bolt.DB
}

var StoreBlockchain, StoreWallet, StoreSettings, StoreMempool *Store

func (store *Store) init() {

	// Open the my.store data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("./_build/"+store.Name+".store", 0600, nil)

	if err != nil {
		gui.Fatal(err)
	}

	store.DB = db

	gui.Log("Store Opened " + store.Name)

}

func (store *Store) close() {
	store.DB.Close()
	gui.Log("Store Closed " + store.Name)
}

func DBInit() {

	var prefix = "network"
	switch config.NETWORK_SELECTED {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix = "main"
	case config.TEST_NET_NETWORK_BYTE:
		prefix = "test"
	case config.DEV_NET_NETWORK_BYTE:
		prefix = "dev"
	}

	if _, err := os.Stat("./_build"); os.IsNotExist(err) {
		if err := os.Mkdir("./_build", 0755); err != nil {
			panic(err)
		}
	}

	if _, err := os.Stat("./_build/" + prefix); os.IsNotExist(err) {
		if err := os.Mkdir("./_build/"+prefix, 0755); err != nil {
			panic(err)
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

	if err := StoreSettings.DB.Update(func(boltTx *bolt.Tx) (err error) {
		_, err = boltTx.CreateBucketIfNotExists([]byte("Settings"))
		return
	}); err != nil {
		panic(err)
	}

	if err := StoreWallet.DB.Update(func(boltTx *bolt.Tx) (err error) {
		_, err = boltTx.CreateBucketIfNotExists([]byte("Wallet"))
		return
	}); err != nil {
		panic(err)
	}

	if err := StoreBlockchain.DB.Update(func(boltTx *bolt.Tx) (err error) {
		if _, err = boltTx.CreateBucketIfNotExists([]byte("Chain")); err != nil {
			return
		}
		if _, err = boltTx.CreateBucketIfNotExists([]byte("Accounts")); err != nil {
			return
		}
		_, err = boltTx.CreateBucketIfNotExists([]byte("Tokens"))
		return
	}); err != nil {
		panic(err)
	}

	if err := StoreMempool.DB.Update(func(boltTx *bolt.Tx) (err error) {
		return
	}); err != nil {
		panic(err)
	}
}

func DBClose() {
	StoreBlockchain.close()
	StoreWallet.close()
	StoreSettings.close()
	StoreMempool.close()
}
