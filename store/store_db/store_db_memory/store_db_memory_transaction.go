package store_db_memory

import (
	"errors"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"sync"
)

type StoreDBMemoryTransactionData struct {
	value     []byte
	operation string
}

type StoreDBMemoryTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	store map[string][]byte
	write bool
	local *sync.Map
}

func (tx *StoreDBMemoryTransaction) IsWritable() bool {
	return tx.write
}

func (tx *StoreDBMemoryTransaction) Put(key string, value []byte) {
	if !tx.write {
		panic("Transaction is not writeable")
	}
	tx.local.Store(key, &StoreDBMemoryTransactionData{value, "put"})
}

func (tx *StoreDBMemoryTransaction) PutClone(key string, value []byte) {
	tx.Put(key, helpers.CloneBytes(value))
}

func (tx *StoreDBMemoryTransaction) Get(key string) []byte {

	out, ok := tx.local.Load(key)
	if ok {
		data := out.(*StoreDBMemoryTransactionData)
		if data.operation == "del" {
			return nil
		}
		return data.value
	}

	resp := tx.store[key]
	tx.local.Store(key, &StoreDBMemoryTransactionData{resp, "get"})
	return resp
}

func (tx *StoreDBMemoryTransaction) GetClone(key string) []byte {
	return helpers.CloneBytes(tx.Get(key))
}

func (tx *StoreDBMemoryTransaction) Exists(key string) bool {
	data := tx.Get(key)
	if data != nil {
		return true
	}
	return false
}

func (tx *StoreDBMemoryTransaction) Delete(key string) {
	if !tx.write {
		panic("Transaction is not writeable")
	}
	tx.local.Store(key, &StoreDBMemoryTransactionData{nil, "del"})
}

func (tx *StoreDBMemoryTransaction) writeTx() error {

	if !tx.write {
		return errors.New("Transaction is not writeable")
	}

	tx.local.Range(func(key, value interface{}) bool {

		data := value.(*StoreDBMemoryTransactionData)

		if data.operation == "del" {
			delete(tx.store, key.(string))
		} else if data.operation == "put" {
			tx.store[key.(string)] = data.value
		}
		return true
	})

	return nil
}
