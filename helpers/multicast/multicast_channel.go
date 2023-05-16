package multicast

import (
	"golang.org/x/exp/slices"
	"pandora-pay/helpers/linked_list"
	"pandora-pay/helpers/recovery"
	"sync"
)

type MulticastChannel[T any] struct {
	listeners           []chan T
	queueBroadcastCn    chan T
	internalBroadcastCn chan T
	count               int
	lock                *sync.RWMutex
}

func (self *MulticastChannel[T]) AddListener() chan T {
	cn := make(chan T)

	self.lock.Lock()
	defer self.lock.Unlock()
	self.listeners = append(self.listeners, cn)
	return cn
}

func (self *MulticastChannel[T]) Broadcast(data T) {
	self.queueBroadcastCn <- data
}

func (self *MulticastChannel[T]) RemoveChannel(channel chan T) bool {

	self.lock.Lock()
	defer self.lock.Unlock()

	x := slices.Index(self.listeners, channel)
	if x != -1 {
		close(channel)
		self.listeners = slices.Delete(self.listeners, x, x+1)
		return true
	}

	return false
}

func (self *MulticastChannel[T]) CloseAll() {

	self.lock.Lock()
	defer self.lock.Unlock()

	for _, cn := range self.listeners {
		close(cn)
	}
	self.listeners = []chan T{}
	close(self.internalBroadcastCn)
}

func (self *MulticastChannel[T]) runQueueBroadcast() {

	linkedList := linked_list.NewLinkedList[T]()

	for {
		if first, ok := linkedList.GetHead(); ok {
			select {
			case data, ok := <-self.queueBroadcastCn:
				if !ok {
					return
				}
				linkedList.Push(data)
			case self.internalBroadcastCn <- first:
				linkedList.PopHead()
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

		self.lock.RLock()
		listeners := slices.Clone(self.listeners)
		self.lock.RUnlock()

		for _, channel := range listeners {
			recovery.Safe(func() { //could be closed
				channel <- data
			})
		}
	}
}

func NewMulticastChannel[T any]() *MulticastChannel[T] {

	multicast := &MulticastChannel[T]{
		[]chan T{},
		make(chan T),
		make(chan T),
		0,
		&sync.RWMutex{},
	}

	go multicast.runInternalBroadcast()
	go multicast.runQueueBroadcast()

	return multicast
}
