package websockets

import (
	"errors"
	"github.com/gorilla/websocket"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/settings"
	"sync"
)

type Websockets struct {
	AllAddresses  sync.Map
	Clients       []*websocket.Conn
	ServerClients []*websocket.Conn
	All           []*websocket.Conn
	sync.RWMutex  `json:"-"`
}

func (socks *Websockets) NewConnection(conn *websocket.Conn, connType bool) error {

	addr := conn.RemoteAddr().String()

	_, exists := socks.AllAddresses.LoadOrStore(addr, conn)
	if exists {
		conn.Close()
		return errors.New("Already connected ")
	}

	conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))

	socks.Lock()
	defer socks.Unlock()

	socks.All = append(socks.All, conn)
	if connType {
		socks.Clients = append(socks.Clients, conn)
	} else {
		socks.ServerClients = append(socks.ServerClients, conn)
	}

	return nil
}

func CreateWebsockets(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Websockets {

	socks := &Websockets{
		AllAddresses:  sync.Map{},
		Clients:       []*websocket.Conn{},
		ServerClients: []*websocket.Conn{},
		All:           []*websocket.Conn{},
	}

	return socks
}
