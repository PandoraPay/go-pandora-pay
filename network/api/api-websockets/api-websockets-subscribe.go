package api_websockets

import (
	"encoding/json"
	api_common "pandora-pay/network/api/api-common"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribeAccount(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	request := &api_common.APIAccountRequest{api_common.APIAccountRequestData{"", nil}, api_common.RETURN_SERIALIZED}
	if err = json.Unmarshal(values, &request); err != nil {
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

	err = conn.Subscriptions.AddSubscription(connection.SUBSCRIPTION_ACCOUNT, publicKeyHash, request.ReturnType)
	return
}

func (api *APIWebsockets) subscribedAccountNotificationReceived(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	var notification *api_common.APISubscriptionNotification
	if err = json.Unmarshal(values, &notification); err != nil {
		return
	}

	api.AccountsChangesSubscriptionNotifications.Broadcast(notification)
	return
}

func (api *APIWebsockets) unsubscribeAccount(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {
	var usubscribeRequest *api_common.APIAccountUnsubscribeRequest
	if err = json.Unmarshal(values, &usubscribeRequest); err != nil {
		return
	}

	publicKeyHash, err := usubscribeRequest.GetPublicKeyHash()
	if err != nil {
		return
	}

	out = []byte{0}
	if conn.Subscriptions.RemoveSubscription(connection.SUBSCRIPTION_ACCOUNT, publicKeyHash) {
		out = []byte{1}
	}
	return
}
