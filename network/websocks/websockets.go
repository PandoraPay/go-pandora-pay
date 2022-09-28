package websocks

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_http"
	"pandora-pay/network/api/api_websockets"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/network/websocks/websock"
	"pandora-pay/recovery"
	"pandora-pay/settings"
	"strconv"
	"sync/atomic"
	"time"
)

type Websockets struct {
	connectedNodes               *connected_nodes.ConnectedNodes
	knownNodes                   *known_nodes.KnownNodes
	ApiWebsockets                *api_websockets.APIWebsockets
	UpdateNewConnectionMulticast *multicast.MulticastChannel[*connection.AdvancedConnection]
	bannedNodes                  *banned_nodes.BannedNodes
	subscriptions                *WebsocketSubscriptions
	api                          *api_http.API
	settings                     *settings.Settings
}

func (websockets *Websockets) GetClients() int64 {
	return atomic.LoadInt64(&websockets.connectedNodes.Clients)
}

func (websockets *Websockets) GetServerSockets() int64 {
	return atomic.LoadInt64(&websockets.connectedNodes.ServerSockets)
}

func (websockets *Websockets) GetFirstSocket() *connection.AdvancedConnection {
	list := websockets.connectedNodes.AllList.Get()
	if len(list) > 0 {
		return list[0]
	}
	return nil
}

func (websockets *Websockets) GetAllSockets() []*connection.AdvancedConnection {
	return websockets.connectedNodes.AllList.Get()
}

func (websockets *Websockets) GetRandomSocket() *connection.AdvancedConnection {
	list := websockets.GetAllSockets()
	if len(list) > 0 {
		index := rand.Intn(len(list))
		return list[index]
	}
	return nil
}

func (websockets *Websockets) Disconnect() int {
	list := websockets.GetAllSockets()
	for _, sock := range list {
		sock.Close()
	}
	return len(list)
}

func (websockets *Websockets) Broadcast(name []byte, data []byte, consensusTypeAccepted map[config.ConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return
	}

	all := websockets.GetAllSockets()

	for i, conn := range all {
		if conn.UUID != exceptSocketUUID && consensusTypeAccepted[conn.Handshake.Consensus] {
			go func(conn *connection.AdvancedConnection, i int) {
				conn.Send(name, data, ctxDuration)
			}(conn, i)
		}
	}

}

func (websockets *Websockets) BroadcastAwaitAnswer(name, data []byte, consensusTypeAccepted map[config.ConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return nil
	}

	all := websockets.GetAllSockets()

	t := time.Now().Unix()
	index := rand.Int()
	gui.GUI.Log("Propagating", index, len(all), string(name), t)

	chans := make(chan *advanced_connection_types.AdvancedConnectionReply, len(all)+1)
	for i, conn := range all {
		if conn.UUID != exceptSocketUUID && consensusTypeAccepted[conn.Handshake.Consensus] {
			go func(conn *connection.AdvancedConnection, i int) {
				answer := conn.SendAwaitAnswer(name, data, ctx, ctxDuration)
				chans <- answer
			}(conn, i)
		} else {
			chans <- nil
		}
	}

	out := make([]*advanced_connection_types.AdvancedConnectionReply, len(all))
	for i := range all {
		out[i] = <-chans
		if out[i] != nil && out[i].Err != nil {
			gui.GUI.Error("Error propagating", index, out[i].Err, len(all), string(name), all[i].RemoteAddr, all[i].UUID, time.Now().Unix()-t)
		}
	}

	return out
}

func (websockets *Websockets) BroadcastJSON(name []byte, data interface{}, consensusTypeAccepted map[config.ConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {
	out, _ := msgpack.Marshal(data)
	websockets.Broadcast(name, out, consensusTypeAccepted, exceptSocketUUID, ctxDuration)
}

func (websockets *Websockets) BroadcastJSONAwaitAnswer(name []byte, data interface{}, consensusTypeAccepted map[config.ConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {
	out, _ := msgpack.Marshal(data)
	return websockets.BroadcastAwaitAnswer(name, out, consensusTypeAccepted, exceptSocketUUID, ctx, ctxDuration)
}
func (websockets *Websockets) closedConnection(conn *connection.AdvancedConnection) {

	if conn.KnownNode != nil {
		websockets.knownNodes.MarkKnownNodeDisconnected(conn.KnownNode)
	}
	websockets.connectedNodes.JustDisconnected(conn)

	conn.InitializedStatusMutex.Lock()

	if conn.InitializedStatus != connection.INITIALIZED_STATUS_INITIALIZED {
		conn.InitializedStatusMutex.Unlock()
		return
	}

	conn.InitializedStatus = connection.INITIALIZED_STATUS_CLOSED
	conn.InitializedStatusMutex.Unlock()

	totalSockets := websockets.connectedNodes.Disconnected(conn)

	if config.SEED_WALLET_NODES_INFO {
		websockets.subscriptions.websocketClosedCn <- conn
	}

	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
}

func (websockets *Websockets) increaseScoreKnownNode(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool {
	return websockets.knownNodes.IncreaseKnownNodeScore(knownNode, delta, isServer)
}

func (websockets *Websockets) NewConnection(c *websock.Conn, remoteAddr string, knownNode *known_node.KnownNodeScored, connectionType bool) (*connection.AdvancedConnection, error) {

	conn, err := connection.NewAdvancedConnection(c, remoteAddr, knownNode, websockets.ApiWebsockets.GetMap, connectionType, websockets.subscriptions.newSubscriptionCn, websockets.subscriptions.removeSubscriptionCn, websockets.closedConnection, websockets.increaseScoreKnownNode)
	if err != nil {
		return nil, err
	}

	if !websockets.connectedNodes.JustConnected(conn, remoteAddr) {
		return nil, errors.New("Already connected")
	}

	recovery.SafeGo(conn.ReadPump)
	recovery.SafeGo(conn.SendPings)

	if knownNode != nil {
		websockets.knownNodes.MarkKnownNodeConnected(knownNode)
		recovery.SafeGo(conn.IncreaseKnownNodeScore)
	}

	if err = websockets.InitializeConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (websockets *Websockets) InitializeConnection(conn *connection.AdvancedConnection) (err error) {

	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	out := conn.SendAwaitAnswer([]byte("handshake"), nil, nil, 0)

	if out.Err != nil {
		return errors.New("Error sending handshake")
	}
	if out.Out == nil {
		return errors.New("Handshake was not received")
	}

	handshakeReceived := &connection.ConnectionHandshake{}
	if err := msgpack.Unmarshal(out.Out, handshakeReceived); err != nil {
		return errors.New("Handshake received was invalid")
	}

	version, err := handshakeReceived.ValidateHandshake()
	if err != nil {
		return errors.New("Handshake is invalid")
	}

	if handshakeReceived.URL != "" && websockets.bannedNodes.IsBanned(handshakeReceived.URL) {
		return errors.New("Socket is banned")
	}

	conn.Handshake = handshakeReceived
	conn.Version = version

	if conn.IsClosed.IsSet() {
		return
	}

	conn.InitializedStatusMutex.Lock()
	conn.InitializedStatus = connection.INITIALIZED_STATUS_INITIALIZED
	conn.InitializedStatusMutex.Unlock()

	totalSockets := websockets.connectedNodes.ConnectedHandshakeValidated(conn)
	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)

	websockets.UpdateNewConnectionMulticast.Broadcast(conn)

	return nil
}

func NewWebsockets(chain *blockchain.Blockchain, mempool *mempool.Mempool, settings *settings.Settings, connectedNodes *connected_nodes.ConnectedNodes, knownNodes *known_nodes.KnownNodes, bannedNodes *banned_nodes.BannedNodes, api *api_http.API, apiWebsockets *api_websockets.APIWebsockets) *Websockets {

	websockets := &Websockets{
		connectedNodes:               connectedNodes,
		knownNodes:                   knownNodes,
		UpdateNewConnectionMulticast: multicast.NewMulticastChannel[*connection.AdvancedConnection](),
		api:                          api,
		ApiWebsockets:                apiWebsockets,
		settings:                     settings,
		bannedNodes:                  bannedNodes,
	}

	websockets.subscriptions = newWebsocketSubscriptions(websockets, chain, mempool)

	recovery.SafeGo(func() {
		for {
			gui.GUI.InfoUpdate("sockets", strconv.FormatInt(atomic.LoadInt64(&connectedNodes.Clients), 32)+" "+strconv.FormatInt(atomic.LoadInt64(&connectedNodes.ServerSockets), 32))
			time.Sleep(1 * time.Second)
		}
	})

	websockets.initializeConsensus(chain, mempool)

	return websockets
}
