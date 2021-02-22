package store

import (
	bolt "go.etcd.io/bbolt"
	"pandora-pay/gui"
)

type Store struct {
	Name   string
	Opened bool
	DB     *bolt.DB
}

var StoreBlockchain = Store{Name: "blockchain"}
var StoreWallet = Store{Name: "wallet"}
var StoreSettings = Store{Name: "settings"}

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

func InitDB() {
	StoreBlockchain.init()
	StoreWallet.init()
	StoreSettings.init()
}

func CloseDB() {
	StoreBlockchain.close()
	StoreWallet.close()
	StoreSettings.close()
}
