package connection

import (
	"context"
	"encoding/json"
	"errors"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/tevino/abool"
	"nhooyr.io/websocket"
	"pandora-pay/config"
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

type AdvancedConnection struct {
	UUID                   string
	Conn                   *websocket.Conn
	Handshake              *ConnectionHandshake
	RemoteAddr             string
	answerCounter          uint32
	Closed                 chan struct{}
	InitializedStatus      InitializedStatusType //use the mutex
	InitializedStatusMutex *sync.Mutex
	IsClosed               *abool.AtomicBool
	getMap                 map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)
	answerMap              map[uint32]chan *AdvancedConnectionAnswer
	answerMapLock          *sync.Mutex
	Subscriptions          *Subscriptions
	ConnectionType         bool
}

func (c *AdvancedConnection) Close(reason string) error {
	if c.IsClosed.SetToIf(false, true) {
		close(c.Closed)
	}
	return c.Conn.Close(websocket.StatusNormalClosure, reason)
}

func (c *AdvancedConnection) connSendJSON(message interface{}) error {

	data, err := json.Marshal(message)
	if err != nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
	defer cancel()

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}

	return c.Conn.Write(ctx, websocket.MessageBinary, data)
}

func (c *AdvancedConnection) connSendPing() error {

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_PONG_WAIT)
	defer cancel()

	if c.IsClosed.IsSet() {
		return errors.New("Closed")
	}
	return c.Conn.Ping(ctx)
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data []byte, reply bool) error {
	message := &AdvancedConnectionMessage{
		replyBackId,
		reply,
		false,
		name,
		data,
	}
	return c.connSendJSON(message)
}

func (c *AdvancedConnection) sendNowAwait(name []byte, data []byte, reply bool) *AdvancedConnectionAnswer {

	replyBackId := atomic.AddUint32(&c.answerCounter, 1)

	eventCn := make(chan *AdvancedConnectionAnswer)
	c.answerMapLock.Lock()
	c.answerMap[replyBackId] = eventCn
	c.answerMapLock.Unlock()

	message := &AdvancedConnectionMessage{
		replyBackId,
		reply,
		true,
		name,
		data,
	}

	if err := c.connSendJSON(message); err != nil {
		return &AdvancedConnectionAnswer{nil, err}
	}

	timer := time.NewTimer(config.WEBSOCKETS_TIMEOUT)
	defer timer.Stop()

	select {
	case out, ok := <-eventCn:
		if !ok {
			return &AdvancedConnectionAnswer{nil, errors.New("Timeout - Closed channel")}
		}
		return out
	case <-timer.C:
		c.answerMapLock.Lock()
		if c.answerMap[replyBackId] != nil {
			delete(c.answerMap, replyBackId)
			close(eventCn)
		}
		c.answerMapLock.Unlock()
		return &AdvancedConnectionAnswer{nil, errors.New("Timeout")}
	}
}

func (c *AdvancedConnection) Send(name []byte, data []byte) error {
	return c.sendNow(0, name, data, false)
}

func (c *AdvancedConnection) SendJSON(name []byte, data interface{}) error {
	out, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.sendNow(0, name, out, false)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data []byte) *AdvancedConnectionAnswer {
	return c.sendNowAwait(name, data, false)
}

func (c *AdvancedConnection) SendJSONAwaitAnswer(name []byte, data interface{}) *AdvancedConnectionAnswer {
	out, err := json.Marshal(data)
	if err != nil {
		panic("Error marshaling data")
	}
	return c.sendNowAwait(name, out, false)
}

func (c *AdvancedConnection) get(message *AdvancedConnectionMessage) ([]byte, error) {

	route := string(message.Name)
	var callback func(conn *AdvancedConnection, values []byte) ([]byte, error)
	if callback = c.getMap[route]; callback != nil {
		return callback(c, message.Data)
	}

	return nil, errors.New("Unknown GET request")
}

func (c *AdvancedConnection) processRead(message *AdvancedConnectionMessage) {

	if !message.ReplyStatus {

		out, err := c.get(message)

		if message.ReplyAwait {
			if err != nil {
				_ = c.sendNow(message.ReplyId, []byte{0}, []byte(err.Error()), true)
			} else {
				_ = c.sendNow(message.ReplyId, []byte{1}, out, true)
			}
		}

	} else {

		output := &AdvancedConnectionAnswer{}
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
			select {
			case cn <- output:
			default:
			}
		}
	}

}

func (c *AdvancedConnection) ReadPump() {

	var cancel context.CancelFunc
	var ctx context.Context

	c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))

	for {

		ctx, cancel = context.WithCancel(context.Background())

		_, read, err := c.Conn.Read(ctx)
		cancel()

		if err != nil {
			c.Close("Timeout read")
			break
		}

		message := new(AdvancedConnectionMessage)
		if err = json.Unmarshal(read, &message); err != nil {
			continue
		}

		//gui.Log(string(message.Name) + " " + strconv.FormatUint(uint64(message.ReplyId), 10) + " " + string(message.Data))

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

	u, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	advancedConnection := &AdvancedConnection{
		UUID:                   u.String(),
		Conn:                   conn,
		Handshake:              nil,
		RemoteAddr:             remoteAddr,
		Closed:                 make(chan struct{}),
		InitializedStatus:      INITIALIZED_STATUS_CREATED,
		InitializedStatusMutex: &sync.Mutex{},
		IsClosed:               abool.New(),
		answerCounter:          0,
		getMap:                 getMap,
		answerMap:              make(map[uint32]chan *AdvancedConnectionAnswer),
		answerMapLock:          &sync.Mutex{},
		ConnectionType:         connectionType,
	}
	advancedConnection.Subscriptions = CreateSubscriptions(advancedConnection, newSubscriptionCn, removeSubscriptionCn)
	return advancedConnection, nil
}
