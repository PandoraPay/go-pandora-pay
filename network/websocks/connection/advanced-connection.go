package connection

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/tevino/abool"
	"pandora-pay/config"
	"sync"
	"sync/atomic"
	"time"
)

type AdvancedConnectionMessage struct {
	ReplyId     uint32
	ReplyStatus bool
	ReplyAwait  bool
	Name        []byte
	Data        []byte
}

type AdvancedConnectionAnswer struct {
	Out []byte
	Err error
}

type AdvancedConnection struct {
	Conn           *websocket.Conn
	answerCounter  uint32
	Closed         chan struct{}
	IsClosed       *abool.AtomicBool
	getMap         map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)
	answerMap      map[uint32]chan *AdvancedConnectionAnswer
	answerMapLock  *sync.RWMutex `json:"-"`
	sendingLock    *sync.Mutex   `json:"-"`
	ConnectionType bool
}

func (c *AdvancedConnection) Close() error {
	if c.IsClosed.SetToIf(false, true) {
		close(c.Closed)
	}
	return c.Conn.Close()
}

func (c *AdvancedConnection) connSendJSON(message interface{}) (err error) {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()
	if err = c.Conn.SetWriteDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
		return
	}
	if err = c.Conn.WriteJSON(message); err != nil {
		return
	}
	return
}

func (c *AdvancedConnection) connSendPing() (err error) {
	c.sendingLock.Lock()
	defer c.sendingLock.Unlock()
	return c.Conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(config.WEBSOCKETS_TIMEOUT))
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

	defer func() {
		c.Close()
	}()

	c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))
	c.Conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_PONG_WAIT))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_PONG_WAIT))
		return nil
	})

	for {

		_, read, err := c.Conn.ReadMessage()
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
		c.Close()
	}()

	for {
		<-pingTicker.C
		if err := c.connSendPing(); err != nil {
			return
		}
	}

}

func CreateAdvancedConnection(conn *websocket.Conn, getMap map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error), connectionType bool) *AdvancedConnection {
	return &AdvancedConnection{
		Conn:           conn,
		Closed:         make(chan struct{}),
		IsClosed:       abool.New(),
		answerCounter:  0,
		getMap:         getMap,
		answerMap:      make(map[uint32]chan *AdvancedConnectionAnswer),
		answerMapLock:  &sync.RWMutex{},
		sendingLock:    &sync.Mutex{},
		ConnectionType: connectionType,
	}
}
