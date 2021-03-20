package websocks

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"pandora-pay/config"
	"pandora-pay/helpers"
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
	apiWebsockets *api.APIWebsockets
	sync.RWMutex  `json:"-"`
}

func (c *AdvancedConnection) sendNow(replyBackId uint32, name []byte, data interface{}, await, reply bool) interface{} {

	if await && replyBackId == 0 {

		c.Lock()
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
		replyBackId = c.answerCounter
		c.answerMap[replyBackId] = make(chan interface{})
		c.Unlock()
	}

	marshal, _ := json.Marshal(data)

	message := &AdvancedConnectionMessage{
		replyBackId,
		reply,
		name,
		marshal,
	}
	c.send <- message
	if await {
		return <-c.answerMap[replyBackId]
	}
	return nil
}

func (c *AdvancedConnection) Send(name []byte, data interface{}) {
	c.sendNow(0, name, data, false, false)
}

func (c *AdvancedConnection) SendAwaitAnswer(name []byte, data interface{}) interface{} {
	return c.sendNow(0, name, data, true, false)
}

func (c *AdvancedConnection) get(message *AdvancedConnectionMessage) (out interface{}, err error) {

	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()

	route := string(message.Name)
	var callback func(values []byte) interface{}
	if callback = c.apiWebsockets.GetMap[route]; callback != nil {
		out = callback(message.Data)
		return
	}

	err = errors.New("Unknown GET request")
	return
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
		if err = c.Conn.ReadJSON(&message); err != nil {
			continue
		}

		if message.Answer == 0 || !message.Reply {

			var out interface{}
			out, err = c.get(message)
			if err != nil {
				c.sendNow(message.Answer, []byte{0}, out, false, true)
			} else {
				c.sendNow(message.Answer, []byte{1}, out, false, true)
			}

		} else {
			c.RLock()
			cn := c.answerMap[message.Answer]
			c.RUnlock()
			if cn != nil {
				cn <- message
			}
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

func CreateAdvancedConnection(conn *websocket.Conn, api *api.API, apiWebsockets *api.APIWebsockets) *AdvancedConnection {
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
