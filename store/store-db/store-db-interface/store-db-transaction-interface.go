package store_db_interface

type StoreDBTransactionInterface interface {
	Put(key string, value []byte) error
	Get(key string) []byte
	Delete(key string) error
	DeleteForcefully(key string) error
}
