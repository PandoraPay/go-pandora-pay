package store_db_bolt

import (
	bolt "go.etcd.io/bbolt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBolt struct {
	store_db_interface.StoreDBInterface
	DB   *bolt.DB
	Name []byte
}

func (store *StoreDBBolt) Close() error {
	return store.DB.Close()
}

func (store *StoreDBBolt) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	return store.DB.View(func(boltTx *bolt.Tx) error {
		tx := StoreDBBoltTransaction{
			boltTx: boltTx,
		}
		return callback(tx)
	})
}

func (store *StoreDBBolt) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	return store.DB.Update(func(boltTx *bolt.Tx) error {
		bucket := boltTx.Bucket(store.Name)
		tx := StoreDBBoltTransaction{
			boltTx: boltTx,
			bucket: bucket,
		}
		return callback(tx)
	})
}

func CreateStoreDBBolt(name string) (store *StoreDBBolt, err error) {

	store = &StoreDBBolt{
		Name: []byte(name),
	}

	// Open the my.store data file in your current directory.
	// It will be created if it doesn't exist.
	if store.DB, err = bolt.Open("./"+name+".store", 0600, nil); err != nil {
		return
	}

	if err = store.DB.Update(func(tx *bolt.Tx) (err error) {
		if _, err = tx.CreateBucketIfNotExists(store.Name); err != nil {
			return
		}
		return
	}); err != nil {
		return
	}

	return
}
