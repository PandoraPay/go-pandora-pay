package multicast

import (
	"pandora-pay/helpers/linked_list"
	"sync"
	"sync/atomic"
)

type MulticastChannel struct {
	listeners           *atomic.Value //[]chan interface{}
	queueBroadcastCn    chan interface{}
	internalBroadcastCn chan interface{}
	count               int
	lock                *sync.Mutex
}

func (self *MulticastChannel) AddListener() <-chan interface{} {
	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load().([]chan interface{})
	newChan := make(chan interface{})

	self.listeners.Store(append(listeners, newChan))
	return newChan

}

func (self *MulticastChannel) Broadcast(data interface{}) {
	self.queueBroadcastCn <- data
}

func (self *MulticastChannel) RemoveChannel(channel <-chan interface{}) bool {

	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load().([]chan interface{})
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

func (self *MulticastChannel) CloseAll() {
	self.lock.Lock()
	defer self.lock.Unlock()

	listeners := self.listeners.Load().([]chan interface{})
	for _, channel := range listeners {
		close(channel)
	}
	self.listeners.Store(make([]chan<- interface{}, 0))

	close(self.internalBroadcastCn)
}

func (self *MulticastChannel) runQueueBroadcast() {

	linkedList := linked_list.NewLinkedList()

	for {
		if first := linkedList.GetFirst(); first != nil {
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

func (self *MulticastChannel) runInternalBroadcast() {

	for {
		data, ok := <-self.internalBroadcastCn
		if !ok {
			return
		}

		listeners := self.listeners.Load().([]chan interface{})
		for _, channel := range listeners {
			channel <- data
		}
	}
}

func NewMulticastChannel() *MulticastChannel {

	multicast := &MulticastChannel{
		&atomic.Value{}, //[]chan interface{}
		make(chan interface{}),
		make(chan interface{}),
		0,
		&sync.Mutex{},
	}

	multicast.listeners.Store(make([]chan interface{}, 0))
	go multicast.runInternalBroadcast()
	go multicast.runQueueBroadcast()

	return multicast
}
