package store_db_in_memory

import store_db_interface "pandora-pay/store/store-db/store-db-interface"

type StoreDBInMemoryTransaction struct {
	store_db_interface.StoreDBTransactionInterface
}

func (tx *StoreDBInMemoryTransaction) Put(key []byte, value []byte) error {
	return nil
}

func (tx *StoreDBInMemoryTransaction) Get(key []byte) []byte {
	return nil
}

func (tx *StoreDBInMemoryTransaction) Delete(key []byte) error {
	return nil
}
