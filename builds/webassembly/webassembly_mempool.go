package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"syscall/js"
)

//NOT USED ANYMORE

func mempoolRemoveTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		hash, err := base64.StdEncoding.DecodeString(args[0].String())
		if err != nil {
			return nil, err
		}

		mutex.Lock()
		defer mutex.Unlock()

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(mempool.CONTINUE_PROCESSING_NO_ERROR_RESET)

		app.Mempool.RemoveInsertedTxsFromBlockchain([]string{string(hash)})

		return nil, nil
	})
}

func mempoolInsertTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		tx := &transaction.Transaction{}
		if err := json.Unmarshal([]byte(args[1].String()), tx); err != nil {
			return nil, err
		}
		if err := tx.BloomAll(); err != nil {
			return nil, err
		}

		err := app.Mempool.AddTxToMempool(tx, 0, false, true, true, advanced_connection_types.UUID_SKIP_ALL, context.Background())

		return nil, err
	})
}
