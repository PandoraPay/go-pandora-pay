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

func DBInit() {
	StoreBlockchain.init()
	StoreWallet.init()
	StoreSettings.init()

	err1 := StoreWallet.DB.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("Wallet"))
		return
	})
	err2 := StoreSettings.DB.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("Settings"))
		return
	})
	err3 := StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {
		if _, err = tx.CreateBucketIfNotExists([]byte("Chain")); err != nil {
			return
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("Accounts")); err != nil {
			return
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Tokens"))
		return
	})

	if err1 != nil || err2 != nil || err3 != nil {
		gui.Log("Wallet bucket creation raised an error")
	}

}

func DBClose() {
	StoreBlockchain.close()
	StoreWallet.close()
	StoreSettings.close()
}
