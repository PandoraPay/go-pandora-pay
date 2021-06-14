package store_db_bolt

import (
	bolt "go.etcd.io/bbolt"
	"os"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

const dbName = "bolt"

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
		tx := &StoreDBBoltTransaction{
			boltTx: boltTx,
			bucket: boltTx.Bucket(store.Name),
		}
		return callback(tx)
	})
}

func (store *StoreDBBolt) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	return store.DB.Update(func(boltTx *bolt.Tx) error {
		tx := &StoreDBBoltTransaction{
			boltTx: boltTx,
			bucket: boltTx.Bucket(store.Name),
		}
		return callback(tx)
	})
}

func CreateStoreDBBolt(name string) (*StoreDBBolt, error) {

	var err error

	store := &StoreDBBolt{
		Name: []byte(name),
	}

	prefix := "./store"
	if _, err = os.Stat(prefix); os.IsNotExist(err) {
		if err = os.Mkdir(prefix, 0755); err != nil {
			return nil, err
		}
	}

	// Open the my.store data file in your current directory.
	// It will be created if it doesn't exist.
	if store.DB, err = bolt.Open(prefix+name+"_store"+"."+dbName, 0600, nil); err != nil {
		return nil, err
	}

	if err = store.DB.Update(func(tx *bolt.Tx) (err error) {
		_, err = tx.CreateBucketIfNotExists(store.Name)
		return
	}); err != nil {
		return nil, err
	}

	return store, nil
}
