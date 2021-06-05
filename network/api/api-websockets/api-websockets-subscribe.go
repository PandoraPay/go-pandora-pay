package api_websockets

import (
	"encoding/json"
	api_common "pandora-pay/network/api/api-common"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribeAccount(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	request := &api_common.APIAccountRequest{"", nil, api_common.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return
	}

	acc, err := api.apiStore.LoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil {
		return
	}

	if acc != nil {
		if request.ReturnType == api_common.RETURN_SERIALIZED {
			out = acc.SerializeToBytes()
		} else {
			if out, err = json.Marshal(acc); err != nil {
				return
			}
		}
	}

	if _, err = conn.Subscriptions.AddSubscription(connection.SUBSCRIPTION_ACCOUNT, publicKeyHash, request.ReturnType); err != nil {
		return
	}

	return
}

func (api *APIWebsockets) subscribeAccountNotification(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	var notification *api_common.APISubscriptionNotification
	if err = json.Unmarshal(values, &notification); err != nil {
		return
	}

	api.AccountsChangesSubscriptionNotifications.Broadcast(notification)

	return
}
