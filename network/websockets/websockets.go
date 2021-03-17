package websockets

import (
	"errors"
	"github.com/gorilla/websocket"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/settings"
	"sync"
)

type Websockets struct {
	allAddresses  sync.Map
	clients       []*websocket.Conn
	serverClients []*websocket.Conn
	all           []*websocket.Conn
	sync.RWMutex  `json:"-"`
}

func (socks *Websockets) NewConnection(conn *websocket.Conn, connType bool) error {

	addr := conn.RemoteAddr().String()

	_, exists := socks.allAddresses.LoadOrStore(addr, conn)
	if exists {
		conn.Close()
		return errors.New("Already connected ")
	}

	socks.Lock()
	defer socks.Unlock()

	socks.all = append(socks.all, conn)
	if connType {
		socks.clients = append(socks.clients, conn)
	} else {
		socks.serverClients = append(socks.serverClients, conn)
	}

	return nil
}

func CreateWebsockets(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Websockets {

	socks := &Websockets{
		clients:       []*websocket.Conn{},
		serverClients: []*websocket.Conn{},
		all:           []*websocket.Conn{},
	}

	return socks
}
