package store_db_interface

type StoreDBTransactionInterface interface {
	Write([]byte, []byte) error
	Get([]byte) []byte
	Delete([]byte) error
}
