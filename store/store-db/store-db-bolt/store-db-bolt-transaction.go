package store_db_bolt

import (
	bolt "go.etcd.io/bbolt"
	"pandora-pay/helpers"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBoltTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	boltTx *bolt.Tx
	bucket *bolt.Bucket
}

func (tx *StoreDBBoltTransaction) Put(key string, value []byte) error {
	return tx.bucket.Put([]byte(key), value)
}

func (tx *StoreDBBoltTransaction) Get(key string) []byte {
	return tx.bucket.Get([]byte(key))
}

func (tx *StoreDBBoltTransaction) GetClone(key string) []byte {
	v := tx.Get(key)
	return helpers.CloneBytes(v)
}

func (tx *StoreDBBoltTransaction) Delete(key string) error {
	return tx.bucket.Delete([]byte(key))
}

func (tx *StoreDBBoltTransaction) DeleteForcefully(key string) error {
	return tx.bucket.Delete([]byte(key))
}
