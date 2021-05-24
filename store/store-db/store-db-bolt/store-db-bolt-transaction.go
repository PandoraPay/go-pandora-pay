package store_db_bolt

import (
	bolt "go.etcd.io/bbolt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBBoltTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	boltTx *bolt.Tx
	bucket *bolt.Bucket
}

func (tx *StoreDBBoltTransaction) Put(key []byte, value []byte) error {
	return tx.bucket.Put(key, value)
}

func (tx *StoreDBBoltTransaction) Get(key []byte) []byte {
	return tx.bucket.Get(key)
}

func (tx *StoreDBBoltTransaction) Delete(key []byte) error {
	return tx.bucket.Delete(key)
}
