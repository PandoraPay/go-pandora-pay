package store_db_interface

type StoreDBTransactionInterface interface {
	Put([]byte, []byte) error
	Get([]byte) []byte
	Delete([]byte) error
}
