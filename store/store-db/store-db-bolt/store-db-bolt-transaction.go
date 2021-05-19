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
