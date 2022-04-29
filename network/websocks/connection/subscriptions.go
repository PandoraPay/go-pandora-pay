package connection

import (
	"bytes"
	"errors"
	"golang.org/x/exp/slices"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/network/api/api_common/api_types"
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

func checkSubscriptionLength(key []byte, subscriptionType api_types.SubscriptionType) error {
	var length int
	switch subscriptionType {
	case api_types.SUBSCRIPTION_PLAIN_ACCOUNT, api_types.SUBSCRIPTION_ACCOUNT, api_types.SUBSCRIPTION_ACCOUNT_TRANSACTIONS:
		length = cryptography.PublicKeyHashSize
	case api_types.SUBSCRIPTION_ASSET:
		length = config_coins.ASSET_LENGTH
	case api_types.SUBSCRIPTION_TRANSACTION:
		length = cryptography.HashSize
	}
	if len(key) != length {
		return errors.New("Key is invalid")
	}
	return nil
}

func (s *Subscriptions) AddSubscription(subscriptionType api_types.SubscriptionType, key []byte, returnType api_types.APIReturnType) error {

	if subscriptionType == api_types.SUBSCRIPTION_PLAIN_ACCOUNT {
		return errors.New("These subscriptions are automatically. They can't be subsribed manually")
	}

	if err := checkSubscriptionLength(key, subscriptionType); err != nil {
		return err
	}

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

func (s *Subscriptions) RemoveSubscription(subscriptionType api_types.SubscriptionType, key []byte) error {

	if err := checkSubscriptionLength(key, subscriptionType); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	for i, subscription := range s.list {
		if subscription.Type == subscriptionType && bytes.Equal(subscription.Key, key) {
			s.list = slices.Delete(s.list, i, i+1)
			s.removeSubscriptionCn <- &SubscriptionNotification{subscription, s.conn}
			return nil
		}
	}

	return errors.New("Subscription not found")
}

func NewSubscriptions(conn *AdvancedConnection, newSubscriptionCn, removeSubscriptionCn chan<- *SubscriptionNotification) (s *Subscriptions) {
	return &Subscriptions{
		conn:                 conn,
		newSubscriptionCn:    newSubscriptionCn,
		removeSubscriptionCn: removeSubscriptionCn,
	}
}
