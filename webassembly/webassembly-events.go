package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config/globals"
	"pandora-pay/helpers"
	"pandora-pay/helpers/events"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/recovery"
	"sync/atomic"
	"syscall/js"
)

func listenEvents(this js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel, id := globals.MainEvents.AddListener()
	defer globals.MainEvents.RemoveChannel(id)

	callback := args[0]
	var err error

	recovery.SafeGo(func() {
		for {
			dataValue, ok := <-channel
			if !ok {
				return
			}

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
	})

	return index
}

func listenNetworkNotifications(this js.Value, args []js.Value) interface{} {
	return normalFunction(func() (interface{}, error) {

		if len(args) != 1 || args[0].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a callback function")
		}
		callback := args[0]

		accountsCn, id := app.Network.Websockets.ApiWebsockets.AccountsChangesSubscriptionNotifications.AddListener()
		defer app.Network.Websockets.ApiWebsockets.AccountsChangesSubscriptionNotifications.RemoveChannel(id)

		recovery.SafeGo(func() {

			var err error
			for {
				dataValue, ok := <-accountsCn
				if !ok {
					return
				}

				data := dataValue.(*api_types.APISubscriptionNotification)

				var acc account.Account
				var output []byte
				if data.Data != nil {
					if err = acc.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
						continue
					}
					if output, err = json.Marshal(acc); err != nil {
						continue
					}
				}

				callback.Invoke(int(api_types.SUBSCRIPTION_ACCOUNT), hex.EncodeToString(data.Key), string(output))
			}
		})

		return true, nil
	})
}
