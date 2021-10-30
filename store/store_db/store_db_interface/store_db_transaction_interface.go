package store_db_interface

type StoreDBTransactionInterface interface {
	Put(key string, value []byte) error
	PutClone(key string, value []byte) error
	Get(key string) []byte
	Exists(key string) bool
	Delete(key string) error
	DeleteForcefully(key string) error
	IsWritable() bool
}
