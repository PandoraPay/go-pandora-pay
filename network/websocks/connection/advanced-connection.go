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
	Conn          *websocket.Conn
	send          chan *AdvancedConnectionMessage
	answerCounter uint32
	Closed        chan struct{}
	IsClosed      *abool.AtomicBool
	getMap        map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)
	answerMap     map[uint32]chan *AdvancedConnectionAnswer
	answerMapLock sync.RWMutex `json:"-"`
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data []byte, await, reply bool) *AdvancedConnectionAnswer {

	if await && replyBackId == 0 {
		replyBackId = atomic.AddUint32(&c.answerCounter, 1)
		c.answerMapLock.Lock()
		c.answerMap[replyBackId] = make(chan *AdvancedConnectionAnswer)
		c.answerMapLock.Unlock()
	}

	message := &AdvancedConnectionMessage{
		replyBackId,
		reply,
		await,
		name,
		data,
	}
	if c.IsClosed.IsSet() {
		return nil
	}
	c.send <- message
	if await {
		select {
		case out, ok := <-c.answerMap[replyBackId]:
			if ok == false {
				return &AdvancedConnectionAnswer{Err: errors.New("Timeout - Closed channel")}
			}
			return out
		case <-time.After(config.WEBSOCKETS_TIMEOUT):
			delete(c.answerMap, replyBackId)
			return &AdvancedConnectionAnswer{Err: errors.New("Timeout")}
		}
	}
	return nil
}

func (c *AdvancedConnection) Send(name []byte, data []byte) {
	c.sendNow(0, name, data, false, false)
}

func (c *AdvancedConnection) SendJSON(name []byte, data interface{}) {
	out, _ := json.Marshal(data)
	c.sendNow(0, name, out, false, false)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data []byte) *AdvancedConnectionAnswer {
	return c.sendNow(0, name, data, true, false)
}

func (c *AdvancedConnection) SendJSONAwaitAnswer(name []byte, data interface{}) *AdvancedConnectionAnswer {
	out, _ := json.Marshal(data)
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

func (c *AdvancedConnection) ReadPump() {

	defer func() {
		if c.IsClosed.SetToIf(false, true) {
			close(c.Closed)
			close(c.send)
		}
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))
	c.Conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_PONG_TIMEOUT))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_PONG_TIMEOUT))
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

		if message.ReplyAwait || !message.ReplyStatus {

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
	}

}

func (c *AdvancedConnection) WritePump() {

	pingTicker := time.NewTicker(config.WEBSOCKETS_PING_TIMEOUT)

	defer func() {
		pingTicker.Stop()
		c.Conn.Close()
		if c.IsClosed.SetToIf(false, true) {
			close(c.Closed)
		}
	}()

	var err error
	for {
		select {
		case message, ok := <-c.send:
			if !ok { // Closed the channel.
				return
			}
			if err = c.Conn.SetWriteDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
				return
			}
			if err = c.Conn.WriteJSON(message); err != nil {
				return
			}
		case <-pingTicker.C:
			if err = c.Conn.SetWriteDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func CreateAdvancedConnection(conn *websocket.Conn, getMap map[string]func(conn *AdvancedConnection, values []byte) ([]byte, error)) *AdvancedConnection {
	return &AdvancedConnection{
		Conn:          conn,
		send:          make(chan *AdvancedConnectionMessage),
		Closed:        make(chan struct{}),
		IsClosed:      abool.New(),
		answerCounter: 0,
		getMap:        getMap,
		answerMap:     make(map[uint32]chan *AdvancedConnectionAnswer),
		answerMapLock: sync.RWMutex{},
	}
}
