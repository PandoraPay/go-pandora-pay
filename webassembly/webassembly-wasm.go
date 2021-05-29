// +build wasm

package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/events"
	"sync/atomic"
	"syscall/js"
)

var subscriptionsIndex uint64
var startMainCallback func()

func SubscribeEvents(none js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()
	callback := args[0]

	go func() {
		for {
			dataValue := <-channel
			data := dataValue.(*events.EventData)

			var final interface{}

			switch v := data.Data.(type) {
			case interface{}:
				str, err := json.Marshal(v)
				if err == nil {
					final = string(str)
				} else {
					final = "error marshaling object"
				}
			default:
				final = data.Data
			}

			callback.Invoke(data.Name, final)
		}
	}()

	return index
}

func HelloPandora(js.Value, []js.Value) interface{} {
	gui.GUI.Info("HelloPandora works!")
	return true
}

func Start(js.Value, []js.Value) interface{} {
	startMainCallback()
	return true
}

func Initialize(startMainCb func()) {

	startMainCallback = startMainCb

	PandoraPayExport := map[string]interface{}{
		"Helpers": js.ValueOf(map[string]interface{}{
			"HelloPandora": js.FuncOf(HelloPandora),
			"Start":        js.FuncOf(Start),
		}),
		"Events": js.ValueOf(map[string]interface{}{
			"Subscribe": js.FuncOf(SubscribeEvents),
		}),
		"Enums": js.ValueOf(map[string]interface{}{
			"Transactions": js.ValueOf(map[string]interface{}{
				"TransactionType": js.ValueOf(map[string]interface{}{
					"TxSimple": js.ValueOf(uint64(transaction_type.TxSimple)),
				}),
				"TransactionSimpleScriptType": js.ValueOf(map[string]interface{}{
					"TxSimpleScriptNormal":   js.ValueOf(uint64(transaction_simple.TxSimpleScriptNormal)),
					"TxSimpleScriptUnstake":  js.ValueOf(uint64(transaction_simple.TxSimpleScriptUnstake)),
					"TxSimpleScriptWithdraw": js.ValueOf(uint64(transaction_simple.TxSimpleScriptWithdraw)),
					"TxSimpleScriptDelegate": js.ValueOf(uint64(transaction_simple.TxSimpleScriptDelegate)),
				}),
			}),
		}),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

}
