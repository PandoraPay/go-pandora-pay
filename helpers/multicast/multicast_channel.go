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
	sync.Mutex
}

func (self *MulticastChannel) AddListener() <-chan interface{} {
	self.Lock()
	defer self.Unlock()

	listeners := self.listeners.Load().([]chan interface{})
	newChan := make(chan interface{})

	self.listeners.Store(append(listeners, newChan))

	return newChan
}

func (self *MulticastChannel) Broadcast(data interface{}) {
	self.queueBroadcastCn <- data
}

func (self *MulticastChannel) RemoveChannel(channel <-chan interface{}) bool {

	self.Lock()
	defer self.Unlock()

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
	self.Lock()
	defer self.Unlock()

	listeners := self.listeners.Load().([]chan interface{})
	for _, channel := range listeners {
		close(channel)
	}
	self.listeners.Store(make([]chan<- interface{}, 0))
}

func (self *MulticastChannel) runQueueBroadcast() {

	linkedList := linked_list.NewLinkedList()

	for {
		if first := linkedList.GetFirst(); first != nil {
			select {
			case data := <-self.queueBroadcastCn:
				linkedList.Push(data)
			case self.internalBroadcastCn <- first:
				linkedList.PopFirst()
			}
		} else {
			select {
			case data := <-self.queueBroadcastCn:
				linkedList.Push(data)
			}
		}
	}

}

func (self *MulticastChannel) runInternalBroadcast() {

	for {
		data := <-self.internalBroadcastCn

		listeners := self.listeners.Load().([]chan interface{})
		for _, channel := range listeners {
			channel <- data
		}
	}
}

func NewMulticastChannel() *MulticastChannel {

	multicast := &MulticastChannel{
		listeners:           &atomic.Value{}, //[]chan interface{}
		queueBroadcastCn:    make(chan interface{}),
		internalBroadcastCn: make(chan interface{}),
	}
	multicast.listeners.Store(make([]chan interface{}, 0))

	go multicast.runQueueBroadcast()
	go multicast.runInternalBroadcast()

	return multicast
}
