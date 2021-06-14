package api_websockets

import (
	"encoding/json"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribe(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APISubscriptionRequest{[]byte{}, api_types.SUBSCRIPTION_ACCOUNT, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.AddSubscription(request.Type, request.Key, request.ReturnType)
}

func (api *APIWebsockets) subscribedNotificationReceived(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	var notification *api_types.APISubscriptionNotification
	if err := json.Unmarshal(values, &notification); err != nil {
		return nil, err
	}

	api.AccountsChangesSubscriptionNotifications.Broadcast(notification)
	return nil, nil
}

func (api *APIWebsockets) unsubscribe(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	var unsubscribeRequest *api_types.APIUnsubscriptionRequest
	if err := json.Unmarshal(values, &unsubscribeRequest); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.RemoveSubscription(unsubscribeRequest.Type, unsubscribeRequest.Key)
}
