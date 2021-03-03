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
var StoreMempool = Store{Name: "mempool"}

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

func DBInit() (err error) {
	StoreBlockchain.init()
	StoreWallet.init()
	StoreSettings.init()

	if err = StoreWallet.DB.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("Wallet"))
		return
	}); err != nil {
		return
	}

	if err = StoreSettings.DB.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists([]byte("Settings"))
		return
	}); err != nil {
		return
	}

	if err = StoreBlockchain.DB.Update(func(tx *bolt.Tx) (err error) {
		if _, err = tx.CreateBucketIfNotExists([]byte("Chain")); err != nil {
			return
		}
		if _, err = tx.CreateBucketIfNotExists([]byte("Accounts")); err != nil {
			return
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Tokens"))
		return
	}); err != nil {
		return
	}

	if err = StoreMempool.DB.Update(func(tx *bolt.Tx) (err error) {

		return
	}); err != nil {
		return
	}

	return
}

func DBClose() {
	StoreBlockchain.close()
	StoreWallet.close()
	StoreSettings.close()
	StoreMempool.close()
}
