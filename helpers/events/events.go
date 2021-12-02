package events

import (
	"pandora-pay/helpers/multicast"
)

type EventData struct {
	Name string
	Data interface{}
}

type Events struct {
	multicast.MulticastChannel
}

func (self *Events) BroadcastEventAwait(name string, data interface{}) {

	finalData := &EventData{
		Name: name,
		Data: data,
	}

	self.Broadcast(finalData)
}

func (self *Events) BroadcastEvent(name string, data interface{}) {

	finalData := &EventData{
		Name: name,
		Data: data,
	}

	self.Broadcast(finalData)
}

func NewEvents() *Events {

	events := &Events{
		*multicast.NewMulticastChannel(false),
	}

	return events
}
