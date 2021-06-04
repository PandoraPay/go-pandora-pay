package connection

type SubscriptionType uint8

const (
	SUBSCRIPTION_ACCOUNT SubscriptionType = iota
	SUBSCRIPTION_TOKEN   SubscriptionType = iota
)

type Subscription struct {
	Type   SubscriptionType
	Id     uint64
	Key    []byte
	Option interface{}
}

type SubscriptionNotification struct {
	Subscription *Subscription
	Conn         *AdvancedConnection
}
