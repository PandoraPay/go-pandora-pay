package multicast

import (
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/linked_list"
	"sync"
)

type MulticastChannel[T any] struct {
	listeners           *generics.Value[[]chan T]
	queueBroadcastCn    chan T
	internalBroadcastCn chan T
	count               int
	lock                *sync.Mutex
}

func (self *MulticastChannel[T]) AddListener() <-chan T {

	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load()
	newChan := make(chan T)

	self.listeners.Store(append(listeners, newChan))
	return newChan

}

func (self *MulticastChannel[T]) Broadcast(data T) {
	self.queueBroadcastCn <- data
}

func (self *MulticastChannel[T]) RemoveChannel(channel <-chan T) bool {

	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load()
	for i, cn := range listeners {
		if cn == channel {
			close(cn)
			listeners = append(listeners[:i], listeners[i+1:]...)
			self.listeners.Store(listeners)
			return true
		}
	}

	return false
}

func (self *MulticastChannel[T]) CloseAll() {

	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load()
	for _, channel := range listeners {
		close(channel)
	}
	self.listeners.Store(make([]chan T, 0))

	close(self.internalBroadcastCn)
}

func (self *MulticastChannel[T]) runQueueBroadcast() {

	linkedList := linked_list.NewLinkedList[T]()

	for {
		if first, ok := linkedList.GetFirst(); ok {
			select {
			case data, ok := <-self.queueBroadcastCn:
				if !ok {
					return
				}
				linkedList.Push(data)
			case self.internalBroadcastCn <- first:
				linkedList.PopFirst()
			}
		} else {
			select {
			case data, ok := <-self.queueBroadcastCn:
				if !ok {
					return
				}
				linkedList.Push(data)
			}
		}
	}

}

func (self *MulticastChannel[T]) runInternalBroadcast() {

	for {
		data, ok := <-self.internalBroadcastCn
		if !ok {
			return
		}

		listeners := self.listeners.Load()
		for _, channel := range listeners {
			channel <- data
		}
	}
}

func NewMulticastChannel[T any]() *MulticastChannel[T] {

	multicast := &MulticastChannel[T]{
		&generics.Value[[]chan T]{},
		make(chan T),
		make(chan T),
		0,
		&sync.Mutex{},
	}

	multicast.listeners.Store(make([]chan T, 0))

	go multicast.runInternalBroadcast()
	go multicast.runQueueBroadcast()

	return multicast
}
