package webassembly

import (
	"encoding/hex"
	"pandora-pay/app"
)

func mempoolRemoveTx(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		app.Mempool.SuspendProcessingCn <- struct{}{}
		defer app.Mempool.ContinueProcessing(false)

		app.Mempool.RemoveInsertedTxsFromBlockchain([]string{args[0].String()})

	})
}
