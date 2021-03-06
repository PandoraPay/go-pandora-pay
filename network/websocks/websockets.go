package websocks

import (
	"encoding/json"
	"errors"
	"nhooyr.io/websocket"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	api_http "pandora-pay/network/api/api-http"
	"pandora-pay/network/api/api-websockets"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/recovery"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type Websockets struct {
	AllAddresses                 *sync.Map
	ApiWebsockets                *api_websockets.APIWebsockets
	allList                      *atomic.Value //[]*connection.AdvancedConnection
	allListMutex                 *sync.Mutex
	clients                      int64 //use atomic
	serverSockets                int64 //use atomic
	totalSockets                 int64 //use atomic
	UpdateNewConnectionMulticast *multicast.MulticastChannel
	subscriptions                *WebsocketSubscriptions
	api                          *api_http.API
}

func (websockets *Websockets) GetClients() int64 {
	return atomic.LoadInt64(&websockets.clients)
}

func (websockets *Websockets) GetServerSockets() int64 {
	return atomic.LoadInt64(&websockets.serverSockets)
}

func (websockets *Websockets) GetFirstSocket() *connection.AdvancedConnection {
	list := websockets.allList.Load().([]*connection.AdvancedConnection)
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func (websockets *Websockets) GetAllSockets() []*connection.AdvancedConnection {
	return websockets.allList.Load().([]*connection.AdvancedConnection)
}

func (websockets *Websockets) Broadcast(name []byte, data []byte, consensusTypeAccepted map[config.ConsensusType]bool) {

	all := websockets.GetAllSockets()
	for _, conn := range all {
		if consensusTypeAccepted[conn.Handshake.Consensus] {
			conn.Send(name, data)
		}
	}

}

func (websockets *Websockets) BroadcastJSON(name []byte, data interface{}, consensusTypeAccepted map[config.ConsensusType]bool) {
	out, _ := json.Marshal(data)
	websockets.Broadcast(name, out, consensusTypeAccepted)
}

func (websockets *Websockets) closedConnectionNow(conn *connection.AdvancedConnection) bool {

	conn.InitializedStatusMutex.Lock()
	defer conn.InitializedStatusMutex.Unlock()

	if conn.InitializedStatus != connection.INITIALIZED_STATUS_INITIALIZED {
		return false
	}

	websockets.AllAddresses.LoadAndDelete(conn.RemoteAddr)

	websockets.allListMutex.Lock()
	all := websockets.allList.Load().([]*connection.AdvancedConnection)
	for i, conn2 := range all {
		if conn2 == conn {
			//removing atomic.Value array
			list2 := make([]*connection.AdvancedConnection, len(all)-1)
			copy(list2, all)
			if len(all) > 1 && i != len(all)-1 {
				list2[i] = all[len(all)-1]
			}
			websockets.allList.Store(list2)
			break
		}
	}
	websockets.allListMutex.Unlock()
	conn.InitializedStatus = connection.INITIALIZED_STATUS_CLOSED

	return true
}

func (websockets *Websockets) closedConnection(conn *connection.AdvancedConnection) {

	<-conn.Closed

	if !websockets.closedConnectionNow(conn) {
		return
	}

	websockets.subscriptions.websocketClosedCn <- conn

	if conn.ConnectionType {
		atomic.AddInt64(&websockets.serverSockets, -1)
	} else {
		atomic.AddInt64(&websockets.clients, -1)
	}
	totalSockets := atomic.AddInt64(&websockets.totalSockets, -1)
	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
}

func (websockets *Websockets) NewConnection(c *websocket.Conn, addr string, connectionType bool) (*connection.AdvancedConnection, error) {

	conn, err := connection.CreateAdvancedConnection(c, addr, websockets.ApiWebsockets.GetMap, connectionType, websockets.subscriptions.newSubscriptionCn, websockets.subscriptions.removeSubscriptionCn)
	if err != nil {
		return nil, err
	}

	if _, exists := websockets.AllAddresses.LoadOrStore(addr, conn); exists {
		return nil, errors.New("Already connected")
	}

	recovery.SafeGo(conn.ReadPump)
	recovery.SafeGo(conn.WritePump)
	recovery.SafeGo(func() { websockets.closedConnection(conn) })

	if err = websockets.InitializeConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (websockets *Websockets) InitializeConnection(conn *connection.AdvancedConnection) error {

	out := conn.SendAwaitAnswer([]byte("handshake"), nil)

	if out.Err != nil {
		conn.Close("Error sending handshake")
		return nil
	}
	if out.Out == nil {
		conn.Close("Handshake was not received")
		return errors.New("Handshake was not received")
	}

	handshakeReceived := new(connection.ConnectionHandshake)
	if err := json.Unmarshal(out.Out, &handshakeReceived); err != nil {
		conn.Close("Handshake received was invalid")
		return errors.New("Handshake received was invalid")
	}

	if err := handshakeReceived.ValidateHandshake(); err != nil {
		conn.Close("Handshake is invalid")
		return errors.New("Handshake is invalid")
	}

	conn.Handshake = handshakeReceived

	conn.InitializedStatusMutex.Lock()
	websockets.allListMutex.Lock()
	websockets.allList.Store(append(websockets.allList.Load().([]*connection.AdvancedConnection), conn))
	websockets.allListMutex.Unlock()
	conn.InitializedStatus = connection.INITIALIZED_STATUS_INITIALIZED
	conn.InitializedStatusMutex.Unlock()

	if conn.ConnectionType {
		atomic.AddInt64(&websockets.serverSockets, +1)
	} else {
		atomic.AddInt64(&websockets.clients, +1)
	}
	totalSockets := atomic.AddInt64(&websockets.totalSockets, +1)

	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)

	websockets.UpdateNewConnectionMulticast.BroadcastAwait(conn)

	return nil
}

func CreateWebsockets(chain *blockchain.Blockchain, api *api_http.API, apiWebsockets *api_websockets.APIWebsockets) *Websockets {

	websockets := &Websockets{
		AllAddresses:                 &sync.Map{},
		clients:                      0,
		serverSockets:                0,
		allList:                      &atomic.Value{}, //[]*connection.AdvancedConnection
		allListMutex:                 &sync.Mutex{},
		UpdateNewConnectionMulticast: multicast.NewMulticastChannel(),
		api:                          api,
		ApiWebsockets:                apiWebsockets,
	}

	websockets.subscriptions = newWebsocketSubscriptions(websockets, chain)
	websockets.allList.Store([]*connection.AdvancedConnection{})

	recovery.SafeGo(func() {
		for {
			gui.GUI.InfoUpdate("sockets", strconv.FormatInt(atomic.LoadInt64(&websockets.clients), 32)+" "+strconv.FormatInt(atomic.LoadInt64(&websockets.serverSockets), 32))
			time.Sleep(1 * time.Second)
		}
	})

	return websockets
}
