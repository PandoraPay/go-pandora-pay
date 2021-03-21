package websocks

import (
	"errors"
	"pandora-pay/gui"
	"pandora-pay/network/api"
	"pandora-pay/network/api-websockets"
	"pandora-pay/network/websocks/connection"
	"strconv"
	"sync"
	"time"
)

type Websockets struct {
	AllAddresses  sync.Map
	All           []*connection.AdvancedConnection
	Clients       int
	ServerClients int
	apiWebsockets *api_websockets.APIWebsockets
	api           *api.API
	sync.RWMutex  `json:"-"`
}

func (websockets *Websockets) Broadcast(name []byte, data interface{}) {
	websockets.RLock()
	all := make([]*connection.AdvancedConnection, len(websockets.All))
	copy(all, websockets.All)
	defer websockets.RUnlock()

	for _, conn := range all {
		conn.Send(name, data)
	}
}

func (websockets *Websockets) closedConnection(conn *connection.AdvancedConnection, connType bool) {

	<-conn.Closed

	addr := conn.Conn.RemoteAddr().String()
	conn2, exists := websockets.AllAddresses.LoadAndDelete(addr)
	if !exists || conn2 != conn {
		return
	}

	websockets.Lock()
	defer websockets.Unlock()
	for i, conn2 := range websockets.All {
		if conn2 == conn {
			websockets.All[i] = websockets.All[len(websockets.All)-1]
			websockets.All = websockets.All[:len(websockets.All)-1]
			break
		}
	}

	if connType {
		websockets.Clients -= 1
	} else {
		websockets.ServerClients -= 1
	}
}

func (websockets *Websockets) NewConnection(conn *connection.AdvancedConnection, connType bool) error {

	addr := conn.Conn.RemoteAddr().String()

	_, exists := websockets.AllAddresses.LoadOrStore(addr, conn)
	if exists {
		conn.Conn.Close()
		return errors.New("Already connected")
	}

	websockets.Lock()
	defer websockets.Unlock()

	websockets.All = append(websockets.All, conn)
	if connType {
		websockets.Clients += 1
	} else {
		websockets.ServerClients += 1
	}

	go conn.ReadPump()
	go conn.WritePump()
	go websockets.closedConnection(conn, connType)

	return nil
}

func CreateWebsockets(api *api.API, apiWebsockets *api_websockets.APIWebsockets) *Websockets {

	websockets := &Websockets{
		AllAddresses:  sync.Map{},
		Clients:       0,
		ServerClients: 0,
		All:           []*connection.AdvancedConnection{},
		api:           api,
		apiWebsockets: apiWebsockets,
	}

	go func() {
		for {
			websockets.RLock()
			gui.InfoUpdate("sockets", strconv.Itoa(websockets.Clients)+" "+strconv.Itoa(websockets.ServerClients))
			websockets.RUnlock()
			time.Sleep(1 * time.Second)
		}
	}()

	return websockets
}
