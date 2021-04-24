package websocks

import (
	"encoding/json"
	"errors"
	"pandora-pay/gui"
	api_http "pandora-pay/network/api/api-http"
	"pandora-pay/network/api/api-websockets"
	"pandora-pay/network/websocks/connection"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Websockets struct {
	AllAddresses sync.Map

	AllList      atomic.Value //[]*connection.AdvancedConnection
	AllListMutex sync.Mutex

	Clients       int64
	ServerClients int64

	apiWebsockets *api_websockets.APIWebsockets
	api           *api_http.API
}

func (websockets *Websockets) GetAllSockets() []*connection.AdvancedConnection {
	return websockets.AllList.Load().([]*connection.AdvancedConnection)
}

func (websockets *Websockets) Broadcast(name []byte, data []byte) {

	all := websockets.GetAllSockets()
	for _, conn := range all {
		conn.Send(name, data)
	}

}

func (websockets *Websockets) BroadcastJSON(name []byte, data interface{}) {
	out, _ := json.Marshal(data)
	websockets.Broadcast(name, out)
}

func (websockets *Websockets) closedConnection(conn *connection.AdvancedConnection) {

	<-conn.Closed

	adr := conn.Conn.RemoteAddr().String()
	conn2, exists := websockets.AllAddresses.LoadAndDelete(adr)
	if !exists || conn2 != conn {
		return
	}

	websockets.AllListMutex.Lock()
	all := websockets.AllList.Load().([]*connection.AdvancedConnection)
	for i, conn2 := range all {
		if conn2 == conn {
			//order is not important
			all[i] = all[len(all)-1]
			all = all[:len(all)-1]
			websockets.AllList.Store(all)
			break
		}
	}
	websockets.AllListMutex.Unlock()

	if conn.ConnectionType {
		atomic.AddInt64(&websockets.ServerClients, -1)
	} else {
		atomic.AddInt64(&websockets.Clients, -1)
	}
}

func (websockets *Websockets) NewConnection(conn *connection.AdvancedConnection) error {

	adr := conn.Conn.RemoteAddr().String()

	_, exists := websockets.AllAddresses.LoadOrStore(adr, conn)
	if exists {
		conn.Conn.Close()
		return errors.New("Already connected")
	}

	websockets.AllListMutex.Lock()
	websockets.AllList.Store(append(websockets.AllList.Load().([]*connection.AdvancedConnection), conn))
	websockets.AllListMutex.Unlock()

	if conn.ConnectionType {
		atomic.AddInt64(&websockets.ServerClients, +1)
	} else {
		atomic.AddInt64(&websockets.Clients, +1)
	}

	go conn.ReadPump()
	go conn.WritePump()
	go websockets.closedConnection(conn)

	return nil
}

func CreateWebsockets(api *api_http.API, apiWebsockets *api_websockets.APIWebsockets) *Websockets {

	websockets := &Websockets{
		AllAddresses:  sync.Map{},
		Clients:       0,
		ServerClients: 0,
		AllList:       atomic.Value{},
		AllListMutex:  sync.Mutex{},
		api:           api,
		apiWebsockets: apiWebsockets,
	}
	websockets.AllList.Store([]*connection.AdvancedConnection{})

	go func() {
		for {
			gui.InfoUpdate("sockets", strconv.FormatInt(atomic.LoadInt64(&websockets.Clients), 32)+" "+strconv.FormatInt(atomic.LoadInt64(&websockets.ServerClients), 32))
			time.Sleep(1 * time.Second)
		}
	}()

	return websockets
}
