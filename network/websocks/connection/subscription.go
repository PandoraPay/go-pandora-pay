package connection

import "pandora-pay/network/api_implementation/api_common/api_types"

type Subscription struct {
	Type       api_types.SubscriptionType
	Key        []byte
	ReturnType api_types.APIReturnType
}

type SubscriptionNotification struct {
	Subscription *Subscription
	Conn         *AdvancedConnection
}
