package store_db_memory

import (
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sync"
)

type StoreDBMemory struct {
	store_db_interface.StoreDBInterface
	Name    []byte
	store   map[string][]byte
	rwmutex *sync.RWMutex
}

func (store *StoreDBMemory) Close() error {
	return nil
}

func (store *StoreDBMemory) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.RLock()
	defer store.rwmutex.RUnlock()

	tx := &StoreDBMemoryTransaction{
		store: store.store,
		local: &sync.Map{},
	}
	return callback(tx)
}

func (store *StoreDBMemory) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.Lock()
	defer store.rwmutex.Unlock()

	tx := &StoreDBMemoryTransaction{
		store: store.store,
		local: &sync.Map{},
		write: true,
	}

	err := callback(tx)

	if err == nil {
		if err = tx.writeTx(); err != nil {
			return err
		}
	}

	return nil
}

func CreateStoreDBMemory(name string) (*StoreDBMemory, error) {
	return &StoreDBMemory{
		Name:    []byte(name),
		store:   make(map[string][]byte),
		rwmutex: &sync.RWMutex{},
	}, nil

}
