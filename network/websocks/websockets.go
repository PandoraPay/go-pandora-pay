package websocks

import (
	"context"
	"errors"
	"github.com/tevino/abool"
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/msgpack"
	"pandora-pay/helpers/multicast"
	"pandora-pay/helpers/recovery"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/network_config"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/network/websocks/websock"
	"pandora-pay/settings"
	"strconv"
	"sync/atomic"
	"time"
)

type SocketEvent struct {
	Type         string
	Conn         *connection.AdvancedConnection
	TotalSockets int64
}

type websocketsType struct {
	apiGetMap                    map[string]func(conn *connection.AdvancedConnection, values []byte) (any, error)
	UpdateNewConnectionMulticast *multicast.MulticastChannel[*connection.AdvancedConnection]
	subscriptions                *WebsocketSubscriptions
	UpdateSocketEventMulticast   *multicast.MulticastChannel[*SocketEvent]
	ReadyCn                      *generics.Value[chan struct{}]
	ReadyCnClosed                *abool.AtomicBool
	settings                     *settings.Settings
}

var Websockets *websocketsType

func (this *websocketsType) GetClients() int64 {
	return atomic.LoadInt64(&connected_nodes.ConnectedNodes.Clients)
}

func (this *websocketsType) GetServerSockets() int64 {
	return atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets)
}

func (this *websocketsType) GetAllSockets() []*connection.AdvancedConnection {
	return connected_nodes.ConnectedNodes.AllList.Get()
}

func (this *websocketsType) GetRandomSocket() *connection.AdvancedConnection {
	list := this.GetAllSockets()
	if len(list) > 0 {
		index := rand.Intn(len(list))
		return list[index]
	}
	return nil
}

func (this *websocketsType) Disconnect() int {
	list := this.GetAllSockets()
	for _, sock := range list {
		sock.Close()
	}
	return len(list)
}

func (this *websocketsType) Broadcast(name []byte, data []byte, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return
	}

	all := this.GetAllSockets()

	for i, conn := range all {
		if conn.UUID != exceptSocketUUID && consensusTypeAccepted[conn.Handshake.Consensus] {
			go func(conn *connection.AdvancedConnection, i int) {
				conn.Send(name, data, ctxDuration)
			}(conn, i)
		}
	}

}

func (this *websocketsType) BroadcastAwaitAnswer(name, data []byte, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {

	if exceptSocketUUID == advanced_connection_types.UUID_SKIP_ALL {
		return nil
	}

	all := this.GetAllSockets()

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

func (this *websocketsType) BroadcastJSON(name []byte, data interface{}, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctxDuration time.Duration) {
	out, _ := msgpack.Marshal(data)
	this.Broadcast(name, out, consensusTypeAccepted, exceptSocketUUID, ctxDuration)
}

func (this *websocketsType) BroadcastJSONAwaitAnswer(name []byte, data interface{}, consensusTypeAccepted map[config.NodeConsensusType]bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context, ctxDuration time.Duration) []*advanced_connection_types.AdvancedConnectionReply {
	out, _ := msgpack.Marshal(data)
	return this.BroadcastAwaitAnswer(name, out, consensusTypeAccepted, exceptSocketUUID, ctx, ctxDuration)
}
func (this *websocketsType) closedConnection(conn *connection.AdvancedConnection) {

	if conn.KnownNode != nil {
		known_nodes.KnownNodes.MarkKnownNodeDisconnected(conn.KnownNode)
	}
	connected_nodes.ConnectedNodes.JustDisconnected(conn)

	conn.InitializedStatusMutex.Lock()

	if conn.InitializedStatus != connection.INITIALIZED_STATUS_INITIALIZED {
		conn.InitializedStatusMutex.Unlock()
		return
	}

	conn.InitializedStatus = connection.INITIALIZED_STATUS_CLOSED
	conn.InitializedStatusMutex.Unlock()

	totalSockets := connected_nodes.ConnectedNodes.Disconnected(conn)

	if network_config.NETWORK_ENABLE_SUBSCRIPTIONS {
		this.subscriptions.websocketClosedCn <- conn
	}

	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
	this.UpdateSocketEventMulticast.Broadcast(&SocketEvent{"disconnected", conn, totalSockets})

	if totalSockets < network_config.NETWORK_CONNECTIONS_READY_THRESHOLD {
		if this.ReadyCnClosed.SetToIf(true, false) {
			this.ReadyCn.Store(make(chan struct{}))
		}
	}
}

func (this *websocketsType) increaseScoreKnownNode(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool {
	return known_nodes.KnownNodes.IncreaseKnownNodeScore(knownNode, delta, isServer)
}

func (this *websocketsType) NewConnection(c *websock.Conn, remoteAddr string, knownNode *known_node.KnownNodeScored, connectionType bool) (*connection.AdvancedConnection, error) {

	conn, err := connection.NewAdvancedConnection(c, remoteAddr, knownNode, this.apiGetMap, connectionType, this.subscriptions.newSubscriptionCn, this.subscriptions.removeSubscriptionCn, this.closedConnection, this.increaseScoreKnownNode)
	if err != nil {
		return nil, err
	}

	if !connected_nodes.ConnectedNodes.JustConnected(conn, remoteAddr) {
		return nil, errors.New("Already connected")
	}

	recovery.SafeGo(conn.ReadPump)
	recovery.SafeGo(conn.SendPings)

	if knownNode != nil {
		known_nodes.KnownNodes.MarkKnownNodeConnected(knownNode)
		recovery.SafeGo(conn.IncreaseKnownNodeScore)
	}

	if err = this.InitializeConnection(conn); err != nil {
		return nil, err
	}

	return conn, nil
}

func (this *websocketsType) InitializeConnection(conn *connection.AdvancedConnection) (err error) {

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

	if handshakeReceived.URL != "" && banned_nodes.BannedNodes.IsBanned(handshakeReceived.URL) {
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

	totalSockets := connected_nodes.ConnectedNodes.ConnectedHandshakeValidated(conn)
	globals.MainEvents.BroadcastEvent("sockets/totalSocketsChanged", totalSockets)
	this.UpdateSocketEventMulticast.Broadcast(&SocketEvent{"connected", conn, totalSockets})
	this.UpdateNewConnectionMulticast.Broadcast(conn)

	if totalSockets >= network_config.NETWORK_CONNECTIONS_READY_THRESHOLD {
		cn := this.ReadyCn.Load()
		if this.ReadyCnClosed.SetToIf(false, true) {
			close(cn)
		}
	}

	return nil
}

func NewWebsockets(chain *blockchain.Blockchain, mempool *mempool.Mempool, settings *settings.Settings, apiGetMap map[string]func(conn *connection.AdvancedConnection, values []byte) (any, error)) *websocketsType {

	Websockets = &websocketsType{
		apiGetMap,
		multicast.NewMulticastChannel[*connection.AdvancedConnection](),
		nil,
		multicast.NewMulticastChannel[*SocketEvent](),
		&generics.Value[chan struct{}]{},
		abool.NewBool(false),
		settings,
	}

	Websockets.ReadyCn.Store(make(chan struct{}))
	Websockets.subscriptions = newWebsocketSubscriptions(chain, mempool)

	recovery.SafeGo(func() {
		for {
			gui.GUI.InfoUpdate("sockets", strconv.FormatInt(atomic.LoadInt64(&connected_nodes.ConnectedNodes.Clients), 32)+" "+strconv.FormatInt(atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets), 32))
			time.Sleep(1 * time.Second)
		}
	})

	return Websockets
}
