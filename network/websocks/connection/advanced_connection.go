package connection

import (
	"context"
	"errors"
	"github.com/blang/semver/v4"
	"github.com/tevino/abool"
	"github.com/vmihailenco/msgpack/v5"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"sync"
	"sync/atomic"
	"time"
)

type InitializedStatusType uint8

const (
	INITIALIZED_STATUS_CREATED InitializedStatusType = iota
	INITIALIZED_STATUS_CLOSED
	INITIALIZED_STATUS_INITIALIZED
)

var uuidGenerator uint32 //use atomic

type AdvancedConnection struct {
	Authenticated            *abool.AtomicBool
	UUID                     advanced_connection_types.UUID
	Conn                     *websocket.Conn
	Handshake                *ConnectionHandshake
	Version                  *semver.Version
	KnownNode                *known_node.KnownNodeScored
	RemoteAddr               string
	answerCounter            uint32
	Closed                   chan struct{}
	InitializedStatus        InitializedStatusType //use the mutex
	InitializedStatusMutex   *sync.Mutex
	IsClosed                 *abool.AtomicBool
	getMap                   map[string]func(conn *AdvancedConnection, values []byte) (interface{}, error)
	answerMap                map[uint32]chan *advanced_connection_types.AdvancedConnectionReply
	answerMapLock            *sync.Mutex
	Subscriptions            *Subscriptions
	ConnectionType           bool
	onClosedConnection       func(c *AdvancedConnection)
	onIncreaseKnownNodeScore func(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool
}

func (c *AdvancedConnection) GetTimeout() time.Duration {
	return config.WEBSOCKETS_TIMEOUT
}

func (c *AdvancedConnection) Close(reason string) error {
	if c.IsClosed.SetToIf(false, true) {
		close(c.Closed)
		c.onClosedConnection(c)
		return c.Conn.Close(websocket.StatusNormalClosure, reason[:generics.Min(100, len(reason))])
	}
	return nil
}

func (c *AdvancedConnection) connSendMessage(message interface{}, ctxParent context.Context, ctxDuration time.Duration) error {

	data, err := msgpack.Marshal(message)
	if err != nil {
		return nil
	}

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}

	ctx, cancel := context.WithTimeout(helpers.GetContext(ctxParent), generics.Max(ctxDuration, config.WEBSOCKETS_TIMEOUT))
	defer cancel()

	return c.Conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data []byte, reply bool, ctxParent context.Context, ctxDuration time.Duration) error {
	message := &advanced_connection_types.AdvancedConnectionMessage{
		replyBackId,
		reply,
		false,
		name,
		data,
	}
	return c.connSendMessage(message, ctxParent, ctxDuration)
}

func (c *AdvancedConnection) sendNowAwait(name []byte, data []byte, reply bool, ctxParent context.Context, ctxDuration time.Duration) *advanced_connection_types.AdvancedConnectionReply {

	ctx, cancel := context.WithTimeout(helpers.GetContext(ctxParent), generics.Max(ctxDuration, config.WEBSOCKETS_TIMEOUT))
	defer cancel()

	replyBackId := atomic.AddUint32(&c.answerCounter, 1)

	eventCn := make(chan *advanced_connection_types.AdvancedConnectionReply)
	c.answerMapLock.Lock()
	c.answerMap[replyBackId] = eventCn
	c.answerMapLock.Unlock()

	message := &advanced_connection_types.AdvancedConnectionMessage{
		replyBackId,
		reply,
		true,
		name,
		data,
	}

	if err := c.connSendMessage(message, ctx, ctxDuration); err != nil {
		return &advanced_connection_types.AdvancedConnectionReply{nil, err}
	}

	select {
	case out := <-eventCn:
		return out
	case <-c.Closed:
		return &advanced_connection_types.AdvancedConnectionReply{nil, errors.New("Timeout Closed")}
	case <-ctx.Done():

		var closeChannel bool

		c.answerMapLock.Lock()
		if c.answerMap[replyBackId] != nil {
			delete(c.answerMap, replyBackId)
			closeChannel = true
		}
		c.answerMapLock.Unlock()

		if closeChannel {
			close(eventCn)
		}

		return &advanced_connection_types.AdvancedConnectionReply{nil, errors.New("Timeout")}
	}
}

func (c *AdvancedConnection) Send(name []byte, data []byte, ctxParent context.Context, ctxDuration time.Duration) error {
	return c.sendNow(0, name, data, false, ctxParent, ctxDuration)
}

func (c *AdvancedConnection) SendJSON(name []byte, data interface{}, ctxParent context.Context, ctxDuration time.Duration) error {
	out, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	return c.sendNow(0, name, out, false, ctxParent, ctxDuration)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data []byte, ctxParent context.Context, ctxDuration time.Duration) *advanced_connection_types.AdvancedConnectionReply {
	return c.sendNowAwait(name, data, false, ctxParent, ctxDuration)
}

func SendJSONAwaitAnswer[T any](c *AdvancedConnection, name []byte, data any, ctxParent context.Context, ctxDuration time.Duration) (*T, error) {
	if c == nil {
		return nil, errors.New("Socket is null")
	}
	input, err := msgpack.Marshal(data)
	if err != nil {
		return nil, errors.New("Error marshaling data")
	}
	out := c.sendNowAwait(name, input, false, ctxParent, ctxDuration)
	if out.Err != nil {
		return nil, out.Err
	}

	final := new(T)
	if err = msgpack.Unmarshal(out.Out, final); err != nil {
		return nil, err
	}
	return final, nil
}

func (c *AdvancedConnection) get(message *advanced_connection_types.AdvancedConnectionMessage) (final []byte, err error) {

	defer func() {
		if err2 := recover(); err2 != nil {
			err = err2.(error)
		}
	}()

	var output interface{}

	route := string(message.Name)
	var callback func(conn *AdvancedConnection, values []byte) (interface{}, error)
	if callback = c.getMap[route]; callback != nil {
		output, err = callback(c, message.Data)
		if err != nil {
			return nil, err
		}
	} else {
		err = errors.New("Unknown request")
	}

	if err != nil {
		return nil, err
	}

	switch v := output.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	default:
		return msgpack.Marshal(output)
	}

}

func (c *AdvancedConnection) processRead(message *advanced_connection_types.AdvancedConnectionMessage) {

	if !message.ReplyStatus {

		out, err := c.get(message)

		if message.ReplyAwait {
			if err != nil {
				_ = c.sendNow(message.ReplyId, []byte{0}, []byte(err.Error()), true, nil, config.WEBSOCKETS_TIMEOUT)
			} else {
				_ = c.sendNow(message.ReplyId, []byte{1}, out, true, nil, config.WEBSOCKETS_TIMEOUT)
			}
		}

	} else {

		output := &advanced_connection_types.AdvancedConnectionReply{}
		if len(message.Name) == 1 && message.Name[0] == 1 {
			output.Out = message.Data
		} else {
			output.Err = errors.New(string(message.Data))
		}

		c.answerMapLock.Lock()
		cn := c.answerMap[message.ReplyId]
		if cn != nil {
			delete(c.answerMap, message.ReplyId)
		}
		c.answerMapLock.Unlock()

		if cn != nil {
			cn <- output
		}
	}
}

func (c *AdvancedConnection) ReadPump() {

	c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {

		_, read, err := c.Conn.Read(ctx)

		if err != nil {
			c.Close("Error reading")
			return
		}

		recovery.SafeGo(func() {
			message := &advanced_connection_types.AdvancedConnectionMessage{}
			if err = msgpack.Unmarshal(read, message); err == nil {
				c.processRead(message)
			}
		})

	}

}

func (c *AdvancedConnection) connSendPing() error {

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_PONG_WAIT)
	defer cancel()

	return c.Conn.Ping(ctx)
}

func (c *AdvancedConnection) SendPings() {

	pingTicker := time.NewTicker(config.WEBSOCKETS_PING_INTERVAL)
	defer pingTicker.Stop()

	for {

		select {
		case <-pingTicker.C:
		case <-c.Closed:
			return
		}

		if err := c.connSendPing(); err != nil {
			c.Close(err.Error())
			return
		}

	}

}

func (c *AdvancedConnection) IncreaseKnownNodeScore() {

	ticker := time.NewTicker(config.WEBSOCKETS_INCREASE_KNOWN_NODE_SCORE_INTERVAL)
	defer ticker.Stop()

	for {

		select {
		case <-ticker.C:
		case <-c.Closed:
			return
		}

		if !c.onIncreaseKnownNodeScore(c.KnownNode, 1, c.ConnectionType) {
			break
		}
	}

}

func NewAdvancedConnection(conn *websocket.Conn, remoteAddr string, knownNode *known_node.KnownNodeScored, getMap map[string]func(conn *AdvancedConnection, values []byte) (interface{}, error), connectionType bool, newSubscriptionCn, removeSubscriptionCn chan<- *SubscriptionNotification, onClosedConnection func(*AdvancedConnection), onIncreaseKnownNodeScore func(*known_node.KnownNodeScored, int32, bool) bool) (*AdvancedConnection, error) {

	//making sure u is not collided with UUID_ALL and UUID_SKIP_ALL
	uuid := advanced_connection_types.UUID(atomic.AddUint32(&uuidGenerator, 1))
	for uuid <= advanced_connection_types.UUID_SKIP_ALL {
		uuid = advanced_connection_types.UUID(atomic.AddUint32(&uuidGenerator, 1))
	}

	advancedConnection := &AdvancedConnection{
		abool.New(),
		uuid,
		conn,
		nil,
		nil,
		knownNode,
		remoteAddr,
		0,
		make(chan struct{}),
		INITIALIZED_STATUS_CREATED,
		&sync.Mutex{},
		abool.New(),
		getMap,
		make(map[uint32]chan *advanced_connection_types.AdvancedConnectionReply),
		&sync.Mutex{},
		nil,
		connectionType,
		onClosedConnection,
		onIncreaseKnownNodeScore,
	}
	advancedConnection.Subscriptions = NewSubscriptions(advancedConnection, newSubscriptionCn, removeSubscriptionCn)
	return advancedConnection, nil
}
