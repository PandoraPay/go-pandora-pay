package store_db_wasm

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type StoreDBWASM struct {
	store_db_interface.StoreDBInterface
	Name []byte
}

func (store *StoreDBWASM) Close() error {
	return nil
}

func (store *StoreDBWASM) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBWASMTransaction{}
	return callback(tx)
}

func (store *StoreDBWASM) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	tx := &StoreDBWASMTransaction{}
	return callback(tx)
}

func CreateStoreDBWASM(name string) (store *StoreDBWASM, err error) {

	store = &StoreDBWASM{
		Name: []byte(name),
	}

	return
}
