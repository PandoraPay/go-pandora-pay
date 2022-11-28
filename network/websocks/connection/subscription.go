package connection

import (
	"pandora-pay/network/api_code/api_code_types"
)

type Subscription struct {
	Type       api_code_types.SubscriptionType
	Key        []byte
	ReturnType api_code_types.APIReturnType
}

type SubscriptionNotification struct {
	Subscription *Subscription
	Conn         *AdvancedConnection
}
