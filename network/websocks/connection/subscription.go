package connection

import api_common "pandora-pay/network/api/api-common"

type SubscriptionType uint8

const (
	SUBSCRIPTION_ACCOUNT SubscriptionType = iota
	SUBSCRIPTION_TOKEN
	SUBSCRIPTION_TRANSACTIONS
)

type Subscription struct {
	Type       SubscriptionType
	Key        []byte
	ReturnType api_common.APIReturnType
}

type SubscriptionNotification struct {
	Subscription *Subscription
	Conn         *AdvancedConnection
}
