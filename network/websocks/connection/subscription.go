package connection

type Subscription struct {
	Id     uint64
	Name   []byte
	Key    []byte
	Option interface{}
}

type SubscriptionNotification struct {
	Subscription *Subscription
	Conn         *AdvancedConnection
}
