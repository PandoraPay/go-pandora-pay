package multicast

import (
	"sync"
	"sync/atomic"
)

type MulticastChannelData struct {
	Answer chan interface{}
	Data   interface{}
}

type MulticastChannel struct {
	listeners *atomic.Value //[]chan interface{}
	count     int
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

	listeners := self.listeners.Load().([]chan interface{})

	for _, channel := range listeners {
		channel <- data
	}

}

func (self *MulticastChannel) BroadcastAwait(data interface{}) {

	listeners := self.listeners.Load().([]chan interface{})

	answers := make(chan interface{}, len(listeners)+1)

	for _, channel := range listeners {

		channel <- &MulticastChannelData{
			answers,
			data,
		}

	}

	for range listeners {
		<-answers
	}

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

func NewMulticastChannel() *MulticastChannel {

	multicast := &MulticastChannel{
		listeners: &atomic.Value{}, //[]chan interface{}
	}
	multicast.listeners.Store(make([]chan interface{}, 0))

	return multicast
}
