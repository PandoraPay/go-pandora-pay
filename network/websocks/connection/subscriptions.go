package connection

import (
	"errors"
	"sync"
)

type Subscriptions struct {
	conn              *AdvancedConnection
	list              []*Subscription
	newSubscriptionCn chan<- *SubscriptionNotification
	index             uint64
	sync.Mutex
}

func (s *Subscriptions) AddSubscription(subscriptionType SubscriptionType, key []byte, option interface{}) (uint64, error) {

	s.Lock()
	defer s.Unlock()

	if len(s.list) > 20 {
		return 0, errors.New("Too many subscriptions")
	}

	s.index += 1
	id := s.index

	subscription := &Subscription{subscriptionType, id, key, option}
	s.list = append(s.list, subscription)

	s.newSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}

	return id, nil
}

func (s *Subscriptions) RemoveSubscription(id uint64) bool {
	s.Lock()
	defer s.Unlock()
	for i, subscription := range s.list {
		if subscription.Id == id {
			s.list = append(s.list[:i], s.list[i+1:]...)
			return true
		}
	}

	return false
}

func CreateSubscriptions(conn *AdvancedConnection, newSubscriptionCn chan<- *SubscriptionNotification) (s *Subscriptions) {
	return &Subscriptions{
		conn:              conn,
		newSubscriptionCn: newSubscriptionCn,
	}
}
