package connection

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	api_common "pandora-pay/network/api/api-common"
	"sync"
)

type Subscriptions struct {
	conn                 *AdvancedConnection
	list                 []*Subscription
	newSubscriptionCn    chan<- *SubscriptionNotification
	removeSubscriptionCn chan<- *SubscriptionNotification
	index                uint64
	sync.Mutex
}

func (s *Subscriptions) AddSubscription(subscriptionType SubscriptionType, key []byte, returnType api_common.APIReturnType) error {

	s.Lock()
	defer s.Unlock()

	if len(s.list) > config.WEBSOCKETS_MAX_SUBSCRIPTIONS {
		return errors.New("Too many subscriptions")
	}

	for _, subscription := range s.list {
		if subscription.Type == subscriptionType && bytes.Equal(subscription.Key, key) {
			return errors.New("Already subscribed")
		}
	}

	s.index += 1

	subscription := &Subscription{subscriptionType, key, returnType}
	s.list = append(s.list, subscription)

	s.newSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}

	return nil
}

func (s *Subscriptions) RemoveSubscription(subscriptionType SubscriptionType, key []byte) bool {

	s.Lock()
	defer s.Unlock()

	for i, subscription := range s.list {
		if subscription.Type == subscriptionType && bytes.Equal(subscription.Key, key) {
			s.list = append(s.list[:i], s.list[i+1:]...)
			s.removeSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}
			return true
		}
	}

	return false
}

func CreateSubscriptions(conn *AdvancedConnection, newSubscriptionCn, removeSubscriptionCn chan<- *SubscriptionNotification) (s *Subscriptions) {
	return &Subscriptions{
		conn:                 conn,
		newSubscriptionCn:    newSubscriptionCn,
		removeSubscriptionCn: removeSubscriptionCn,
	}
}
