package store_db_js

import (
	"errors"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
	"sync"
	"syscall/js"
)

type StoreDBJS struct {
	store_db_interface.StoreDBInterface
	Name    []byte
	jsStore js.Value
	rwmutex *sync.RWMutex
}

func (store *StoreDBJS) Close() error {
	return nil
}

func (store *StoreDBJS) View(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.RLock()
	defer store.rwmutex.RUnlock()

	tx := &StoreDBJSTransaction{
		jsStore: store.jsStore,
		local:   &sync.Map{},
	}
	return callback(tx)
}

func (store *StoreDBJS) Update(callback func(dbTx store_db_interface.StoreDBTransactionInterface) error) error {
	store.rwmutex.Lock()
	defer store.rwmutex.Unlock()

	tx := &StoreDBJSTransaction{
		jsStore: store.jsStore,
		local:   &sync.Map{},
		write:   true,
	}

	err := callback(tx)

	if err == nil {
		if err = tx.writeTx(); err != nil {
			return err
		}
	}

	return nil
}

func CreateStoreDBJS(name string) (*StoreDBJS, error) {

	pandoraStorage := js.Global().Get("PandoraStorage")
	if pandoraStorage.IsNull() || pandoraStorage.IsUndefined() {
		return nil, errors.New("`global.PandoraStorage` is missing")
	}

	out := pandoraStorage.Call("createStore", name)
	if out.IsNull() || out.IsUndefined() {
		return nil, errors.New("`createStore` returned a null value")
	}

	return &StoreDBJS{
		Name:    []byte(name),
		jsStore: out,
		rwmutex: &sync.RWMutex{},
	}, nil

}
