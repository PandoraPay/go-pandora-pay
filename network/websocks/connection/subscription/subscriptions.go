package subscription

import "sync"

type Subscriptions struct {
	list  []*Subscription
	index uint64
	sync.Mutex
}

func (s *Subscriptions) AddSubscription(name, key []byte, option interface{}) uint64 {

	s.Lock()
	defer s.Unlock()

	s.index += 1
	id := s.index

	subscription := &Subscription{id, name, key, option}
	s.list = append(s.list, subscription)

	return id
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
