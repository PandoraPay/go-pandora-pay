package websocks

import (
	"errors"
	"pandora-pay/network/api"
	"sync"
)

type Websockets struct {
	AllAddresses  sync.Map
	Clients       []*AdvancedConnection
	ServerClients []*AdvancedConnection
	All           []*AdvancedConnection
	apiWebsockets *api.APIWebsockets
	api           *api.API
	sync.RWMutex  `json:"-"`
}

func (websockets *Websockets) NewConnection(conn *AdvancedConnection, connType bool) error {

	addr := conn.Conn.RemoteAddr().String()

	_, exists := websockets.AllAddresses.LoadOrStore(addr, conn)
	if exists {
		conn.Conn.Close()
		return errors.New("Already connected ")
	}

	websockets.Lock()
	defer websockets.Unlock()

	websockets.All = append(websockets.All, conn)
	if connType {
		websockets.Clients = append(websockets.Clients, conn)
	} else {
		websockets.ServerClients = append(websockets.ServerClients, conn)
	}

	go conn.readPump()
	go conn.writePump()

	return nil
}

func CreateWebsockets(api *api.API, apiWebsockets *api.APIWebsockets) *Websockets {

	websockets := &Websockets{
		AllAddresses:  sync.Map{},
		Clients:       []*AdvancedConnection{},
		ServerClients: []*AdvancedConnection{},
		All:           []*AdvancedConnection{},
		api:           api,
		apiWebsockets: apiWebsockets,
	}

	return websockets
}
