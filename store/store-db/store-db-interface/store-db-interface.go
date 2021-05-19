package store_db_interface

type StoreDBInterface interface {
	Close() error
	View(callback func(dbTx StoreDBTransactionInterface) error) error
	Update(callback func(dbTx StoreDBTransactionInterface) error) error
}
