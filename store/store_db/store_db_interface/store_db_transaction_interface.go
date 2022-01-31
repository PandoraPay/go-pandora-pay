package store_db_interface

type StoreDBTransactionInterface interface {
	Put(key string, value []byte)
	Get(key string) []byte
	Exists(key string) bool
	Delete(key string)
	IsWritable() bool
}
