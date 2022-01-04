package events

import (
	"pandora-pay/helpers/multicast"
)

type EventData[T any] struct {
	Name string
	Data T
}

type Events[T any] struct {
	*multicast.MulticastChannel[*EventData[T]]
}

func (self *Events[T]) BroadcastEventAwait(name string, data T) {

	finalData := &EventData[T]{
		Name: name,
		Data: data,
	}

	self.Broadcast(finalData)
}

func (self *Events[T]) BroadcastEvent(name string, data T) {

	finalData := &EventData[T]{
		Name: name,
		Data: data,
	}

	self.Broadcast(finalData)
}

func NewEvents[T any]() *Events[T] {

	events := &Events[T]{
		multicast.NewMulticastChannel[*EventData[T]](),
	}

	return events
}
