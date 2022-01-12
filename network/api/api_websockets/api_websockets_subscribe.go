package api_websockets

import (
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	request := &api_types.APISubscriptionRequest{[]byte{}, api_types.SUBSCRIPTION_ACCOUNT, api_types.RETURN_SERIALIZED}
	if err := msgpack.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.AddSubscription(request.Type, request.Key, request.ReturnType)
}

func (api *APIWebsockets) subscribedNotificationReceived(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	notification := &api_types.APISubscriptionNotification{}
	if err := msgpack.Unmarshal(values, notification); err != nil {
		return nil, err
	}

	api.SubscriptionNotifications.Broadcast(notification)

	return nil, nil
}

func (api *APIWebsockets) unsubscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	unsubscribeRequest := &api_types.APIUnsubscriptionRequest{}
	if err := msgpack.Unmarshal(values, unsubscribeRequest); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.RemoveSubscription(unsubscribeRequest.Type, unsubscribeRequest.Key)
}
