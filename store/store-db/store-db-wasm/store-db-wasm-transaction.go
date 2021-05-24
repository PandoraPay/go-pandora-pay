package store_db_wasm

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBWASMTransaction struct {
	store_db_interface.StoreDBTransactionInterface
}

func (tx *StoreDBWASMTransaction) Put(key []byte, value []byte) error {
	return nil
}

func (tx *StoreDBWASMTransaction) Get(key []byte) []byte {
	return nil
}

func (tx *StoreDBWASMTransaction) Delete(key []byte) error {
	return nil
}
