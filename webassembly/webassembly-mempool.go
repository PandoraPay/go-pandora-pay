package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction"
	"syscall/js"
)

func mempoolRemoveTx(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		hash, err := hex.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(false)

		app.Mempool.RemoveInsertedTxsFromBlockchain([]string{string(hash)})

		return nil, nil
	})
}

func mempoolInsertTx(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		tx := &transaction.Transaction{}
		if err := json.Unmarshal([]byte(args[1].String()), tx); err != nil {
			return nil, err
		}

		out := app.Mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, 0, true, false, "*")

		return out, nil
	})
}
