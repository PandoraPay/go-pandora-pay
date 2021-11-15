package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/globals"
	"pandora-pay/helpers"
	"pandora-pay/helpers/events"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/recovery"
	"pandora-pay/webassembly/webassembly_utils"
	"sync/atomic"
	"syscall/js"
)

func listenEvents(this js.Value, args []js.Value) interface{} {

	if len(args) == 0 || args[0].Type() != js.TypeFunction {
		return errors.New("Argument must be a callback")
	}

	index := atomic.AddUint64(&subscriptionsIndex, 1)
	channel := globals.MainEvents.AddListener()

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
				if final, err = webassembly_utils.ConvertJSONBytes(v); err != nil {
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
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 1 || args[0].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a callback function")
		}
		callback := args[0]

		subscriptionsCn := app.Network.Websockets.ApiWebsockets.SubscriptionNotifications.AddListener()

		recovery.SafeGo(func() {

			defer app.Network.Websockets.ApiWebsockets.SubscriptionNotifications.RemoveChannel(subscriptionsCn)

			var err error
			for {
				dataValue, ok := <-subscriptionsCn
				if !ok {
					return
				}

				data := dataValue.(*api_types.APISubscriptionNotification)

				var object interface{}

				//gui.GUI.Log(int(data.SubscriptionType))

				switch data.SubscriptionType {
				case api_types.SUBSCRIPTION_ACCOUNT:
					var acc *account.Account
					if data.Data != nil {

						if acc, err = account.NewAccount(data.Key, 0, nil); err != nil {
							continue
						}
						if err = acc.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = acc
				case api_types.SUBSCRIPTION_PLAIN_ACCOUNT:
					plainAcc := plain_account.NewPlainAccount(data.Key, 0)
					if data.Data != nil {
						if err = plainAcc.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = plainAcc
				case api_types.SUBSCRIPTION_ASSET:
					ast := asset.NewAsset(data.Key, 0)
					if data.Data != nil {
						if err = ast.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = ast
				case api_types.SUBSCRIPTION_REGISTRATION:
					reg := registration.NewRegistration(data.Key, 0)
					if data.Data != nil {
						if err = reg.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = reg
				case api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS:
					object = data.Data
				case api_types.SUBSCRIPTION_TRANSACTION:
					object = data.Data
				}

				var output []byte
				if output, err = json.Marshal(object); err != nil {
					continue
				}

				jsOutData := js.Global().Get("Uint8Array").New(len(output))
				js.CopyBytesToJS(jsOutData, output)

				jsOutExtra := js.Global().Get("Uint8Array").New(len(data.Extra))
				js.CopyBytesToJS(jsOutExtra, data.Extra)

				callback.Invoke(int(data.SubscriptionType), hex.EncodeToString(data.Key), jsOutData, jsOutExtra)

			}
		})

		return true, nil
	})
}
