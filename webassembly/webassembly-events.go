package webassembly

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/blockchain/data/accounts/account"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations/registration"
	"pandora-pay/blockchain/data/tokens/token"
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
	return promiseFunction(func() (interface{}, error) {

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
					acc := account.NewAccount(data.Key, nil)
					if data.Data != nil {
						if err = acc.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = acc
				case api_types.SUBSCRIPTION_PLAIN_ACCOUNT:
					plainAcc := new(plain_account.PlainAccount)
					if data.Data != nil {
						if err = plainAcc.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					plainAcc.PublicKey = data.Key
					object = plainAcc
				case api_types.SUBSCRIPTION_TOKEN:
					tok := new(token.Token)
					if data.Data != nil {
						if err = tok.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					object = tok
				case api_types.SUBSCRIPTION_REGISTRATION:
					reg := new(registration.Registration)
					if data.Data != nil {
						if err = reg.Deserialize(helpers.NewBufferReader(data.Data)); err != nil {
							continue
						}
					}
					reg.PublicKey = data.Key
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

				callback.Invoke(int(data.SubscriptionType), hex.EncodeToString(data.Key), string(output), string(data.Extra))

			}
		})

		return true, nil
	})
}
