package connection

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/tevino/abool"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"sync"
	"sync/atomic"
	"time"
)

type AdvancedConnection struct {
	Conn           *websocket.Conn
	Handshake      *ConnectionHandshake
	Initialized    bool
	RemoteAddr     string
	answerCounter  uint32
	Closed         chan struct{}
	IsClosed       *abool.AtomicBool
	getMap         map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)
	answerMap      map[uint32]chan *AdvancedConnectionAnswer
	answerMapLock  *sync.RWMutex
	Subscriptions  *Subscriptions
	ConnectionType bool
}

func (c *AdvancedConnection) Close(reason string) error {
	if c.IsClosed.SetToIf(false, true) {
		close(c.Closed)
	}
	return c.Conn.Close(websocket.StatusNormalClosure, reason)
}

func (c *AdvancedConnection) connSendJSON(message interface{}) (err error) {

	var data []byte
	if data, err = json.Marshal(message); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
	defer cancel()

	err = c.Conn.Write(ctx, websocket.MessageBinary, data)
	return
}

func (c *AdvancedConnection) connSendPing() (err error) {

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_PONG_WAIT)
	defer cancel()

	err = c.Conn.Ping(ctx)
	return
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data []byte, await, reply bool) *AdvancedConnectionAnswer {

	var eventCn chan *AdvancedConnectionAnswer
	if await {
		if replyBackId == 0 {
			replyBackId = atomic.AddUint32(&c.answerCounter, 1)
			eventCn = make(chan *AdvancedConnectionAnswer)
			c.answerMapLock.Lock()
			c.answerMap[replyBackId] = eventCn
			c.answerMapLock.Unlock()
		} else {
			c.answerMapLock.RLock()
			eventCn = c.answerMap[replyBackId]
			c.answerMapLock.RUnlock()
		}
	}

	message := &AdvancedConnectionMessage{
		replyBackId,
		reply,
		await,
		name,
		data,
	}
	if c.IsClosed.IsSet() {
		return &AdvancedConnectionAnswer{nil, errors.New("Closed")}
	}

	// gui.Log(string(message.Name) + " " + strconv.FormatUint(uint64(message.ReplyId), 10) + " " + string(message.Data))

	if err := c.connSendJSON(message); err != nil {
		return &AdvancedConnectionAnswer{nil, err}
	}

	if await {
		timer := time.NewTimer(config.WEBSOCKETS_TIMEOUT)
		select {
		case out, ok := <-eventCn:
			timer.Stop()
			if !ok {
				return &AdvancedConnectionAnswer{nil, errors.New("Timeout - Closed channel")}
			}
			return out
		case <-timer.C:
			c.answerMapLock.Lock()
			delete(c.answerMap, replyBackId)
			c.answerMapLock.Unlock()
			return &AdvancedConnectionAnswer{nil, errors.New("Timeout")}
		}
	}
	return &AdvancedConnectionAnswer{nil, nil}
}

func (c *AdvancedConnection) Send(name []byte, data []byte) {
	c.sendNow(0, name, data, false, false)
}

func (c *AdvancedConnection) SendJSON(name []byte, data interface{}) {
	out, err := json.Marshal(data)
	if err != nil {
		panic("Error marshaling data")
	}
	c.sendNow(0, name, out, false, false)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data []byte) *AdvancedConnectionAnswer {
	return c.sendNow(0, name, data, true, false)
}

func (c *AdvancedConnection) SendJSONAwaitAnswer(name []byte, data interface{}) *AdvancedConnectionAnswer {
	out, err := json.Marshal(data)
	if err != nil {
		panic("Error marshaling data")
	}
	return c.sendNow(0, name, out, true, false)
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
				marshalErr, _ := json.Marshal(err)
				c.sendNow(message.ReplyId, []byte{0}, marshalErr, false, true)
			} else {
				c.sendNow(message.ReplyId, []byte{1}, out, false, true)
			}
		}

	} else {

		output := &AdvancedConnectionAnswer{}
		if bytes.Equal(message.Name, []byte{1}) {
			output.Out = message.Data
		} else {
			if err := json.Unmarshal(message.Data, &output.Err); err != nil {
				output.Err = errors.New("Error decoding received error")
			}
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

	var cancel context.CancelFunc
	var ctx context.Context

	defer func() {
		if cancel != nil {
			cancel()
		}
		c.Close("Timeout read")
	}()

	c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))

	for {

		ctx, cancel = context.WithCancel(context.Background())

		_, read, err := c.Conn.Read(ctx)
		if err != nil {
			break
		}

		message := new(AdvancedConnectionMessage)
		if err = json.Unmarshal(read, &message); err != nil {
			continue
		}

		//gui.Log(string(message.Name) + " " + strconv.FormatUint(uint64(message.ReplyId), 10) + " " + string(message.Data))

		go func() {

			if !message.ReplyStatus {

				var out []byte
				out, err = c.get(message)

				if message.ReplyAwait {
					if err != nil {
						marshalErr, _ := json.Marshal(err)
						c.sendNow(message.ReplyId, []byte{0}, marshalErr, false, true)
					} else {
						c.sendNow(message.ReplyId, []byte{1}, out, false, true)
					}
				}

			} else {

				output := &AdvancedConnectionAnswer{}
				if bytes.Equal(message.Name, []byte{1}) {
					output.Out = message.Data
				} else {
					if err = json.Unmarshal(message.Data, &output.Err); err != nil {
						output.Err = errors.New("Error decoding received error")
					}
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

		}()

	}

}

func (c *AdvancedConnection) WritePump() {

	pingTicker := time.NewTicker(config.WEBSOCKETS_PING_INTERVAL)

	defer func() {
		pingTicker.Stop()
		c.Close("Ping send")
	}()

	for {
		_, ok := <-pingTicker.C
		if !ok {
			return
		}

		if err := c.connSendPing(); err != nil {
			return
		}
	}

}

func CreateAdvancedConnection(conn *websocket.Conn, remoteAddr string, getMap map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error), connectionType bool) (advancedConnection *AdvancedConnection) {
	advancedConnection = &AdvancedConnection{
		Conn:           conn,
		Handshake:      nil,
		RemoteAddr:     remoteAddr,
		Closed:         make(chan struct{}),
		IsClosed:       abool.New(),
		answerCounter:  0,
		getMap:         getMap,
		answerMap:      make(map[uint32]chan *AdvancedConnectionAnswer),
		answerMapLock:  &sync.RWMutex{},
		ConnectionType: connectionType,
	}
	advancedConnection.Subscriptions = CreateSubscriptions(advancedConnection)
	return
}
