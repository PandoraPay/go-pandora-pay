package websocks

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"pandora-pay/config"
	"pandora-pay/network/api"
	"sync"
	"time"
)

type AdvancedConnectionMessage struct {
	Answer uint32
	Reply  bool
	Name   []byte
	Data   []byte
}

type AdvancedConnection struct {
	Conn          *websocket.Conn
	send          chan interface{}
	received      chan interface{}
	answerCounter uint32
	answerMap     map[uint32]chan interface{}
	api           *api.API
	apiWebsockets *APIWebsockets
	sync.RWMutex  `json:"-"`
}

func (c *AdvancedConnection) sendNow(name []byte, data interface{}, await bool) interface{} {
	c.Lock()
	marshal, _ := json.Marshal(data)
	c.answerCounter += 1
	if c.answerCounter == 0 {
		c.answerCounter = 1
	}
	for c.answerMap[c.answerCounter] != nil {
		c.answerCounter += 1
		if c.answerCounter == 0 {
			c.answerCounter = 1
		}
	}
	id := c.answerCounter
	c.answerMap[id] = make(chan interface{})
	c.Unlock()

	message := &AdvancedConnectionMessage{
		c.answerCounter,
		false,
		name,
		marshal,
	}
	c.send <- message
	if await {
		return <-c.answerMap[id]
	}
	return nil
}

func (c *AdvancedConnection) Send(name []byte, data interface{}) {
	c.sendNow(name, data, false)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data interface{}) interface{} {
	return c.sendNow(name, data, true)
}

func (c *AdvancedConnection) readPump() {

	defer func() {
		c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		c.Conn.Close()
	}()

	var err error
	for {
		c.Conn.SetReadLimit(int64(config.WEBSOCKETS_MAX_READ))
		message := new(AdvancedConnectionMessage)
		err = c.Conn.ReadJSON(&message)
		if err != nil {
			return
		}

		if message.Answer == 0 || !message.Reply {
			c.received <- message
		} else {
			c.RLock()
			cn := c.answerMap[message.Answer]
			if cn != nil {
				cn <- message
			}
			c.RUnlock()
		}
	}

}

func (c *AdvancedConnection) writePump() {
	defer func() {
		c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
		c.Conn.Close()
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

		}
	}

}

func CreateAdvancedConnection(conn *websocket.Conn, api *api.API, apiWebsockets *APIWebsockets) *AdvancedConnection {
	return &AdvancedConnection{
		Conn:          conn,
		send:          make(chan interface{}),
		received:      make(chan interface{}),
		answerCounter: 0,
		answerMap:     make(map[uint32]chan interface{}),
		api:           api,
		apiWebsockets: apiWebsockets,
	}
}
