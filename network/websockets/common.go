package websockets

import (
	"github.com/gorilla/websocket"
	"pandora-pay/config"
	"time"
)

func Send(conn *websocket.Conn, messageType int, data []byte) (err error) {
	if err = conn.SetWriteDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
		return
	}
	if err = conn.WriteMessage(messageType, data); err != nil {
		return
	}
}
func SendJSON(conn *websocket.Conn, data interface{}) (err error) {
	if err = conn.SetWriteDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
		return
	}
	if err = conn.WriteJSON(data); err != nil {
		return
	}
}
func Read(conn *websocket.Conn) (messageType int, data []byte, err error) {
	if err = conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
		return
	}
	messageType, data, err = conn.ReadMessage()
	return
}
func ReadJSON(conn *websocket.Conn, data interface{}) (err error) {
	if err = conn.SetReadDeadline(time.Now().Add(config.WEBSOCKETS_TIMEOUT)); err != nil {
		return err
	}
	return conn.ReadJSON(data)
}
