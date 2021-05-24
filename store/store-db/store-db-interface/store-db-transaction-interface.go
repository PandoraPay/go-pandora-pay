package store_db_interface

type StoreDBTransactionInterface interface {
	Put(key []byte, value []byte) error
	Get(key []byte) []byte
	Delete(key []byte) error
}
