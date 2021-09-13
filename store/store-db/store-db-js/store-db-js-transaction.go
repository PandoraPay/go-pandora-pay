package store_db_js

import (
	"errors"
	"fmt"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sync"
	"syscall/js"
)

type StoreDBJSTransactionData struct {
	value     []byte
	operation string
}

type StoreDBJSTransaction struct {
	store_db_interface.StoreDBTransactionInterface
	jsStore js.Value
	write   bool
	local   *sync.Map
}

func (tx *StoreDBJSTransaction) Put(key string, value []byte) error {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	tx.local.Store(key, &StoreDBJSTransactionData{value, "put"})
	return nil
}

func (tx *StoreDBJSTransaction) Get(key string) []byte {

	out, ok := tx.local.Load(key)
	if ok {
		data := out.(*StoreDBJSTransactionData)
		if data.operation == "del" {
			return nil
		}
		return data.value
	}

	respCh := make(chan []byte)
	defer close(respCh)

	errCh := make(chan error)
	defer close(errCh)

	promise := tx.jsStore.Call("getItem", key)

	promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		var result []byte
		if !args[0].IsNull() && !args[0].IsUndefined() {
			result = make([]byte, args[0].Get("length").Int())
			js.CopyBytesToGo(result, args[0])
		}

		respCh <- result
		return nil
	}), js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		errCh <- fmt.Errorf("error reading js db %s", args[0].Get("message").String())
		return nil
	}))

	select {
	case resp := <-respCh:
		tx.local.Store(key, &StoreDBJSTransactionData{resp, "get"})
		return resp
	case <-errCh:
		return nil
	}
}

func (tx *StoreDBJSTransaction) Exists(key string) bool {
	result := tx.Get(key)
	if result != nil {
		return true
	}
	return false
}

func (tx *StoreDBJSTransaction) GetClone(key string) []byte {
	return tx.Get(key) //not required
}

func (tx *StoreDBJSTransaction) Delete(key string) error {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	tx.local.Store(key, &StoreDBJSTransactionData{nil, "del"})
	return nil
}

func (tx *StoreDBJSTransaction) DeleteForcefully(key string) (err error) {
	if !tx.write {
		return errors.New("Transaction is not writeable")
	}
	tx.local.Store(key, &StoreDBJSTransactionData{nil, "del"})
	return
}

func (tx *StoreDBJSTransaction) writeTx() error {

	if !tx.write {
		return errors.New("Transaction is not writeable")
	}

	tx.local.Range(func(key, value interface{}) bool {

		data := value.(*StoreDBJSTransactionData)
		respCh := make(chan bool)
		defer close(respCh)

		errCh := make(chan error)
		defer close(errCh)

		process := true
		if data.operation == "del" {

			promise := tx.jsStore.Call("removeItem", key.(string))

			promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				respCh <- true
				return nil
			}), js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				errCh <- fmt.Errorf("error deleting from js db %s", args[0].Get("message").String())
				return nil
			}))

		} else if data.operation == "put" {

			var final js.Value
			if data.value == nil {
				final = js.Null()
			} else {
				final = js.Global().Get("Uint8Array").New(len(data.value))
				js.CopyBytesToJS(final, data.value)
			}

			promise := tx.jsStore.Call("setItem", key.(string), final)

			promise.Call("then", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				respCh <- true
				return nil
			}), js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				errCh <- fmt.Errorf("error writing to js db %s", args[0].Get("message").String())
				return nil
			}))

		} else {
			process = false
		}

		if process {
			select {
			case <-respCh:
			case <-errCh:
			}
		}

		return true
	})

	return nil
}
