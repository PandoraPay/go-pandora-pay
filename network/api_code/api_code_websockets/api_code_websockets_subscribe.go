package api_code_websockets

import (
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/network/api_implementation/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

func Subscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	request := &api_types.APISubscriptionRequest{[]byte{}, api_types.SUBSCRIPTION_ACCOUNT, api_types.RETURN_SERIALIZED}
	if err := msgpack.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.AddSubscription(request.Type, request.Key, request.ReturnType)
}

func SubscribedNotificationReceived(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	notification := &api_types.APISubscriptionNotification{}
	if err := msgpack.Unmarshal(values, notification); err != nil {
		return nil, err
	}

	SubscriptionNotifications.Broadcast(notification)

	return nil, nil
}

func Unsubscribe(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	unsubscribeRequest := &api_types.APIUnsubscriptionRequest{}
	if err := msgpack.Unmarshal(values, unsubscribeRequest); err != nil {
		return nil, err
	}

	return nil, conn.Subscriptions.RemoveSubscription(unsubscribeRequest.Type, unsubscribeRequest.Key)
}
