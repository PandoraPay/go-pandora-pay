package store_db_in_memory

import store_db_interface "pandora-pay/store/store-db/store-db-interface"

type StoreDBInMemory struct {
	store_db_interface.StoreDBInterface
	Name []byte
}

func (store *StoreDBInMemory) Close() error {
	return nil
}

func (store *StoreDBInMemory) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBInMemoryTransaction{}
	return callback(tx)
}

func (store *StoreDBInMemory) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBInMemoryTransaction{}
	return callback(tx)
}

func CreateStoreInMemory(name string) (store *StoreDBInMemory, err error) {

	store = &StoreDBInMemory{
		Name: []byte(name),
	}

	return
}
