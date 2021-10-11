package store_db_bolt

import (
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
)

type StoreDBBoltTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	boltTx *bolt.Tx
	bucket *bolt.Bucket
	write  bool
}

func (tx *StoreDBBoltTransaction) IsWritable() bool {
	return tx.write
}

func (tx *StoreDBBoltTransaction) Put(key string, value []byte) error {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	return tx.bucket.Put([]byte(key), value)
}

func (tx *StoreDBBoltTransaction) Get(key string) []byte {
	return tx.bucket.Get([]byte(key))
}

func (tx *StoreDBBoltTransaction) Exists(key string) bool {
	return tx.bucket.Get([]byte(key)) != nil
}

func (tx *StoreDBBoltTransaction) GetClone(key string) []byte {
	return helpers.CloneBytes(tx.Get(key))
}

func (tx *StoreDBBoltTransaction) PutClone(key string, value []byte) error {
	return tx.Put(key, helpers.CloneBytes(value))
}

func (tx *StoreDBBoltTransaction) Delete(key string) error {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	return tx.bucket.Delete([]byte(key))
}

func (tx *StoreDBBoltTransaction) DeleteForcefully(key string) error {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	return tx.bucket.Delete([]byte(key))
}
