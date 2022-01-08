package multicast

import (
	"pandora-pay/helpers/container_list"
	"pandora-pay/helpers/linked_list"
)

type MulticastChannel[T any] struct {
	listeners           *container_list.ContainerList[chan T]
	queueBroadcastCn    chan T
	internalBroadcastCn chan T
	count               int
}

func (self *MulticastChannel[T]) AddListener() chan T {
	return self.listeners.Push(make(chan T))
}

func (self *MulticastChannel[T]) Broadcast(data T) {
	self.queueBroadcastCn <- data
}

func (self *MulticastChannel[T]) RemoveChannel(channel chan T) bool {

	if self.listeners.Remove(channel) {
		close(channel)
		return true
	}

	return false
}

func (self *MulticastChannel[T]) CloseAll() {

	list := self.listeners.RemoveAll()
	for _, channel := range list {
		close(channel)
	}
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

		listeners := self.listeners.Get()
		for _, channel := range listeners {
			channel <- data
		}
	}
}

func NewMulticastChannel[T any]() *MulticastChannel[T] {

	multicast := &MulticastChannel[T]{
		container_list.NewContainerList[chan T](),
		make(chan T),
		make(chan T),
		0,
	}

	go multicast.runInternalBroadcast()
	go multicast.runQueueBroadcast()

	return multicast
}
