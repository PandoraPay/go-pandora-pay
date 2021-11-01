package api_websockets

import (
	"encoding/json"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribe(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APISubscriptionRequest{[]byte{}, api_types.SubscriptionType_SUBSCRIPTION_ACCOUNT, api_types.APIReturnType_RETURN_SERIALIZED}
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

	api.SubscriptionNotifications.Broadcast(notification)

	return nil, nil
}

func (api *APIWebsockets) unsubscribe(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	var unsubscribeRequest *api_types.APIUnsubscriptionRequest
	if err := json.Unmarshal(values, &unsubscribeRequest); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.RemoveSubscription(unsubscribeRequest.Type, unsubscribeRequest.Key)
}
