package store_db_bolt

import (
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

func (tx *StoreDBBoltTransaction) Put(key string, value []byte) {
	if err := tx.bucket.Put([]byte(key), value); err != nil {
		panic(err)
	}
}

//bolt requires the data to be cloned
//github issue see here https://github.com/etcd-io/bbolt/issues/298
func (tx *StoreDBBoltTransaction) Get(key string) []byte {
	return helpers.CloneBytes(tx.bucket.Get([]byte(key)))
}

func (tx *StoreDBBoltTransaction) GetClone(key string) []byte {
	return tx.Get(key) //it is already cloned
}

func (tx *StoreDBBoltTransaction) Exists(key string) bool {
	return tx.bucket.Get([]byte(key)) != nil
}

func (tx *StoreDBBoltTransaction) PutClone(key string, value []byte) {
	//verified, the value is being cloned by bbolt
	tx.Put(key, value)
}

func (tx *StoreDBBoltTransaction) Delete(key string) {
	tx.bucket.Delete([]byte(key))
}
