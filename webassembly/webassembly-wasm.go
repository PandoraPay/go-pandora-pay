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
	"pandora-pay/wallet"
	"pandora-pay/wallet/address"
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
		"helpers": js.ValueOf(map[string]interface{}{
			"helloPandora": js.FuncOf(HelloPandora),
			"start":        js.FuncOf(Start),
		}),
		"events": js.ValueOf(map[string]interface{}{
			"subscribe": js.FuncOf(SubscribeEvents),
		}),
		"enums": js.ValueOf(map[string]interface{}{
			"transactions": js.ValueOf(map[string]interface{}{
				"transactionType": js.ValueOf(map[string]interface{}{
					"txSimple": js.ValueOf(uint64(transaction_type.TxSimple)),
				}),
				"transactionSimple": js.ValueOf(map[string]interface{}{
					"scriptType": js.ValueOf(map[string]interface{}{
						"scriptNormal":   js.ValueOf(uint64(transaction_simple.ScriptNormal)),
						"scriptUnstake":  js.ValueOf(uint64(transaction_simple.ScriptUnstake)),
						"scriptWithdraw": js.ValueOf(uint64(transaction_simple.ScriptWithdraw)),
						"scriptDelegate": js.ValueOf(uint64(transaction_simple.ScriptDelegate)),
					}),
				}),
			}),
			"wallet": js.ValueOf(map[string]interface{}{
				"version": js.ValueOf(map[string]interface{}{
					"versionSimple": js.ValueOf(int(wallet.VersionSimple)),
				}),
				"encryptedVersion": js.ValueOf(map[string]interface{}{
					"encryptedVersionPlainText": js.ValueOf(int(wallet.EncryptedVersionPlainText)),
					"encryptedVersionEncrypted": js.ValueOf(int(wallet.EncryptedVersionEncryption)),
				}),
				"address": js.ValueOf(map[string]interface{}{
					"version": js.ValueOf(map[string]interface{}{
						"versionTransparent": js.ValueOf(int(wallet_address.VersionTransparent)),
					}),
				}),
			}),
		}),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

}
