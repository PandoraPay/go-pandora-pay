package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/helpers/events"
	"pandora-pay/wallet"
	"pandora-pay/wallet/address"
	"sync/atomic"
	"syscall/js"
)

var subscriptionsIndex uint64
var startMainCallback func()

var promiseConstructor, errorConstructor js.Value

func convertJSON(obj interface{}) (string, error) {

	str, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}

	return string(str), nil
}

func promiseFunction(callback func() (interface{}, error)) interface{} {
	return promiseConstructor.New(js.FuncOf(func(this2 js.Value, args2 []js.Value) interface{} {
		go func() {
			result, err := callback()
			if err != nil {
				args2[1].Invoke(errorConstructor.New(err.Error()))
				return
			}
			args2[0].Invoke(result)
		}()
		return nil
	}))
}

func normalFunction(callback func() (interface{}, error)) interface{} {
	result, err := callback()
	if err != nil {
		return errorConstructor.New(err.Error())
	}
	return result
}

func subscribeEvents(this js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()
	callback := args[0]
	var err error

	go func() {
		for {
			dataValue := <-channel
			data := dataValue.(*events.EventData)

			var final interface{}

			switch v := data.Data.(type) {
			case string:
				final = data.Data
			case interface{}:
				if final, err = convertJSON(v); err != nil {
					panic(err)
				}
			default:
				final = data.Data
			}

			callback.Invoke(data.Name, final)
		}
	}()

	return index
}

func Initialize(startMainCb func()) {

	startMainCallback = startMainCb

	promiseConstructor = js.Global().Get("Promise")
	errorConstructor = js.Global().Get("Error")

	PandoraPayExport := map[string]interface{}{
		"helpers": js.ValueOf(map[string]interface{}{
			"helloPandora": js.FuncOf(helloPandora),
			"start":        js.FuncOf(start),
			"getIdenticon": js.FuncOf(getIdenticon),
		}),
		"events": js.ValueOf(map[string]interface{}{
			"subscribe": js.FuncOf(subscribeEvents),
		}),
		"wallet": js.ValueOf(map[string]interface{}{
			"getWallet": js.FuncOf(getWallet),
			"manager": js.ValueOf(map[string]interface{}{
				"getWalletAddress":        js.FuncOf(getWalletAddress),
				"addNewWalletAddress":     js.FuncOf(addNewWalletAddress),
				"removeWalletAddress":     js.FuncOf(removeWalletAddress),
				"importWalletPrivateKey":  js.FuncOf(importWalletPrivateKey),
				"importWalletJSON":        js.FuncOf(importWalletJSON),
				"importWalletAddressJSON": js.FuncOf(importWalletAddressJSON),
			}),
		}),
		"addresses": js.ValueOf(map[string]interface{}{
			"decodeAddress":   js.FuncOf(decodeAddress),
			"generateAddress": js.FuncOf(generateAddress),
		}),
		"cryptography": js.ValueOf(map[string]interface{}{
			"computePublicKeyHash": js.FuncOf(computePublicKeyHash),
		}),
		"network": js.ValueOf(map[string]interface{}{
			"getNetworkBlockInfo":     js.FuncOf(getNetworkBlockInfo),
			"getNetworkBlockComplete": js.FuncOf(getNetworkBlockComplete),
			"getNetworkTransaction":   js.FuncOf(getNetworkTransaction),
			"subscribeNetworkAccount": js.FuncOf(subscribeNetworkAccount),
		}),
		"enums": js.ValueOf(map[string]interface{}{
			"transactions": js.ValueOf(map[string]interface{}{
				"transactionType": js.ValueOf(map[string]interface{}{
					"TX_SIMPLE": js.ValueOf(uint64(transaction_type.TX_SIMPLE)),
				}),
				"transactionSimple": js.ValueOf(map[string]interface{}{
					"scriptType": js.ValueOf(map[string]interface{}{
						"SCRIPT_NORMAL":   js.ValueOf(uint64(transaction_simple.SCRIPT_NORMAL)),
						"SCRIPT_UNSTAKE":  js.ValueOf(uint64(transaction_simple.SCRIPT_UNSTAKE)),
						"SCRIPT_WITHDRAW": js.ValueOf(uint64(transaction_simple.SCRIPT_WITHDRAW)),
						"SCRIPT_DELEGATE": js.ValueOf(uint64(transaction_simple.SCRIPT_DELEGATE)),
					}),
				}),
			}),
			"wallet": js.ValueOf(map[string]interface{}{
				"version": js.ValueOf(map[string]interface{}{
					"VERSION_SIMPLE": js.ValueOf(int(wallet.VERSION_SIMPLE)),
				}),
				"encryptedVersion": js.ValueOf(map[string]interface{}{
					"ENCRYPTED_VERSION_PLAIN_TEXT": js.ValueOf(int(wallet.ENCRYPTED_VERSION_PLAIN_TEXT)),
					"ENCRYPTED_VERSION_ENCRYPTION": js.ValueOf(int(wallet.ENCRYPTED_VERSION_ENCRYPTION)),
				}),
				"address": js.ValueOf(map[string]interface{}{
					"version": js.ValueOf(map[string]interface{}{
						"VERSION_TRANSPARENT": js.ValueOf(int(wallet_address.VERSION_TRANSPARENT)),
					}),
				}),
			}),
		}),
		"config": js.ValueOf(map[string]interface{}{
			"NAME":                    js.ValueOf(config.NAME),
			"NETWORK_SELECTED":        js.ValueOf(config.NETWORK_SELECTED),
			"NETWORK_SELECTED_NAME":   js.ValueOf(config.NETWORK_SELECTED_NAME),
			"NETWORK_SELECTED_PREFIX": js.ValueOf(config.NETWORK_SELECTED_BYTE_PREFIX),
			"CONSENSUS":               js.ValueOf(uint8(config.CONSENSUS)),
			"coins": js.ValueOf(map[string]interface{}{
				"DECIMAL_SEPARATOR":        js.ValueOf(config.DECIMAL_SEPARATOR),
				"COIN_DENOMINATION":        js.ValueOf(config.COIN_DENOMINATION),
				"COIN_DENOMINATION_FLOAT":  js.ValueOf(config.COIN_DENOMINATION_FLOAT),
				"MAX_SUPPLY_COINS":         js.ValueOf(config.MAX_SUPPLY_COINS),
				"TOKEN_LENGTH":             js.ValueOf(config.TOKEN_LENGTH),
				"NATIVE_TOKEN_NAME":        js.ValueOf(config.NATIVE_TOKEN_NAME),
				"NATIVE_TOKEN_TICKER":      js.ValueOf(config.NATIVE_TOKEN_TICKER),
				"NATIVE_TOKEN_DESCRIPTION": js.ValueOf(config.NATIVE_TOKEN_DESCRIPTION),
				"NATIVE_TOKEN_STRING":      js.ValueOf(config.NATIVE_TOKEN_STRING),
				"convertToUnitsUint64":     js.FuncOf(convertToUnitsUint64),
				"convertToUnits":           js.FuncOf(convertToUnits),
				"convertToBase":            js.FuncOf(convertToBase),
			}),
			"reward": js.ValueOf(map[string]interface{}{
				"getRewardAt": js.FuncOf(getRewardAt),
			}),
		}),
	}

	js.Global().Set("PandoraPay", js.ValueOf(PandoraPayExport))

}
