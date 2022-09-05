package connected_nodes

import (
	"pandora-pay/helpers/container_list"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

type ConnectedNodes struct {
	AllAddresses  *generics.Map[string, *connection.AdvancedConnection]
	AllList       *container_list.ContainerList[*connection.AdvancedConnection]
	Clients       int64 //use atomic
	ServerSockets int64 //use atomic
	TotalSockets  int64 //use atomic
}

func (this *ConnectedNodes) JustConnected(c *connection.AdvancedConnection, remoteAddr string) bool {
	if _, ok := this.AllAddresses.LoadOrStore(remoteAddr, c); !ok {
		return true
	}
	return false
}

func (this *ConnectedNodes) JustDisconnected(c *connection.AdvancedConnection) {
	this.AllAddresses.LoadAndDelete(c.RemoteAddr)
}

func (this *ConnectedNodes) ConnectedHandshakeValidated(c *connection.AdvancedConnection) int64 {
	this.AllList.Push(c)
	if c.ConnectionType {
		atomic.AddInt64(&this.ServerSockets, +1)
	} else {
		atomic.AddInt64(&this.Clients, +1)
	}
	return atomic.AddInt64(&this.TotalSockets, +1)
}

func (this *ConnectedNodes) Disconnected(c *connection.AdvancedConnection) int64 {
	this.AllList.Remove(c)
	if c.ConnectionType {
		atomic.AddInt64(&this.ServerSockets, -1)
	} else {
		atomic.AddInt64(&this.Clients, -1)
	}
	return atomic.AddInt64(&this.TotalSockets, -1)
}

func NewConnectedNodes() *ConnectedNodes {
	return &ConnectedNodes{
		&generics.Map[string, *connection.AdvancedConnection]{},
		container_list.NewContainerList[*connection.AdvancedConnection](),
		0,
		0,
		0,
	}
}
