package connection

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/tevino/abool"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/network/api/api_common/api_types"
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
	UUID                    advanced_connection_types.UUID
	Conn                    *websocket.Conn
	Handshake               *ConnectionHandshake
	RemoteAddr              string
	answerCounter           uint32
	Closed                  chan struct{}
	InitializedStatus       InitializedStatusType //use the mutex
	InitializedStatusMutex  *sync.Mutex
	IsClosed                *abool.AtomicBool
	getMap                  map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)
	answerMap               map[uint32]chan *advanced_connection_types.AdvancedConnectionAnswer
	answerMapLock           *sync.Mutex
	contextConnection       context.Context
	contextConnectionCancel context.CancelFunc
	Subscriptions           *Subscriptions
	ConnectionType          bool
	CommunicationType       api_types.APIReturnType
}

func (c *AdvancedConnection) GetTimeout() time.Duration {
	return config.WEBSOCKETS_TIMEOUT
}

func (c *AdvancedConnection) Close(reason string) error {
	if c.IsClosed.SetToIf(false, true) {
		close(c.Closed)
	}
	return c.Conn.Close(websocket.StatusNormalClosure, reason)
}

func (c *AdvancedConnection) connSendJSON(message interface{}, ctx context.Context) error {

	data, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(context.Background())
		defer cancel()
	}

	return c.Conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *AdvancedConnection) connSendPing() error {

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_PONG_WAIT)
	defer cancel()

	return c.Conn.Ping(ctx)
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data []byte, reply bool, ctx context.Context) error {
	message := &advanced_connection_types.AdvancedConnectionMessage{
		replyBackId,
		reply,
		false,
		name,
		data,
	}
	return c.connSendJSON(message, ctx)
}

func (c *AdvancedConnection) sendNowAwait(name []byte, data []byte, reply bool, ctx context.Context) *advanced_connection_types.AdvancedConnectionAnswer {

	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), c.GetTimeout())
		defer cancel()
	}

	replyBackId := atomic.AddUint32(&c.answerCounter, 1)

	eventCn := make(chan *advanced_connection_types.AdvancedConnectionAnswer)
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

	if err := c.connSendJSON(message, ctx); err != nil {
		return &advanced_connection_types.AdvancedConnectionAnswer{nil, err}
	}

	select {
	case out, ok := <-eventCn:
		if !ok {
			return &advanced_connection_types.AdvancedConnectionAnswer{nil, errors.New("Timeout - Closed channel")}
		}
		return out
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
		return &advanced_connection_types.AdvancedConnectionAnswer{nil, errors.New("Timeout")}
	}
}

func (c *AdvancedConnection) Send(name []byte, data []byte, ctx context.Context) error {
	return c.sendNow(0, name, data, false, ctx)
}

func (c *AdvancedConnection) SendJSON(name []byte, data interface{}, ctx context.Context) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.sendNow(0, name, out, false, ctx)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data []byte, ctx context.Context) *advanced_connection_types.AdvancedConnectionAnswer {
	return c.sendNowAwait(name, data, false, ctx)
}

func (c *AdvancedConnection) SendJSONAwaitAnswer(name []byte, data interface{}, ctx context.Context) *advanced_connection_types.AdvancedConnectionAnswer {
	if c == nil {
		return &advanced_connection_types.AdvancedConnectionAnswer{nil, errors.New("Socket is null")}
	}
	out, err := json.Marshal(data)
	if err != nil {
		return &advanced_connection_types.AdvancedConnectionAnswer{nil, errors.New("Error marshaling data")}
	}
	return c.sendNowAwait(name, out, false, ctx)
}

func (c *AdvancedConnection) get(message *advanced_connection_types.AdvancedConnectionMessage) ([]byte, error) {

	route := string(message.Name)
	var callback func(conn *AdvancedConnection, values []byte) ([]byte, error)
	if callback = c.getMap[route]; callback != nil {
		return callback(c, message.Data)
	}

	return nil, errors.New("Unknown GET request")
}

func (c *AdvancedConnection) processRead(message *advanced_connection_types.AdvancedConnectionMessage) {

	if !message.ReplyStatus {

		out, err := c.get(message)

		if message.ReplyAwait {
			if err != nil {
				_ = c.sendNow(message.ReplyId, []byte{0}, []byte(err.Error()), true, nil)
			} else {
				_ = c.sendNow(message.ReplyId, []byte{1}, out, true, nil)
			}
		}

	} else {

		output := &advanced_connection_types.AdvancedConnectionAnswer{}
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

	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	for {

		_, read, err := c.Conn.Read(ctx)

		if err != nil {
			c.Close("Error reading")
			break
		}

		message := new(advanced_connection_types.AdvancedConnectionMessage)
		if err = json.Unmarshal(read, &message); err != nil {
			continue
		}

		recovery.SafeGo(func() { c.processRead(message) })
	}

}

func (c *AdvancedConnection) WritePump() {

	pingTicker := time.NewTicker(config.WEBSOCKETS_PING_INTERVAL)

	for {

		if _, ok := <-pingTicker.C; !ok {
			break
		}

		if err := c.connSendPing(); err != nil {
			break
		}
	}

	pingTicker.Stop()
	c.Close("Ping send")

}

func CreateAdvancedConnection(conn *websocket.Conn, remoteAddr string, getMap map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error), connectionType bool, newSubscriptionCn, removeSubscriptionCn chan<- *SubscriptionNotification) (*AdvancedConnection, error) {

	u := advanced_connection_types.UUID(0)
	for u <= advanced_connection_types.UUID_SKIP_ALL {
		u = advanced_connection_types.UUID(atomic.AddUint32(&uuidGenerator, 1))
	}

	ctx, cancel := context.WithCancel(context.Background())

	advancedConnection := &AdvancedConnection{
		UUID:                    u,
		Conn:                    conn,
		Handshake:               nil,
		RemoteAddr:              remoteAddr,
		Closed:                  make(chan struct{}),
		InitializedStatus:       INITIALIZED_STATUS_CREATED,
		InitializedStatusMutex:  &sync.Mutex{},
		IsClosed:                abool.New(),
		answerCounter:           0,
		getMap:                  getMap,
		answerMap:               make(map[uint32]chan *advanced_connection_types.AdvancedConnectionAnswer),
		answerMapLock:           &sync.Mutex{},
		ConnectionType:          connectionType,
		contextConnection:       ctx,
		contextConnectionCancel: cancel,
	}
	advancedConnection.Subscriptions = CreateSubscriptions(advancedConnection, newSubscriptionCn, removeSubscriptionCn)
	return advancedConnection, nil
}
