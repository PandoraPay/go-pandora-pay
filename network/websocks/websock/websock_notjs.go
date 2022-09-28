//go:build !js
// +build !js

package websock

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Conn struct {
	*websocket.Conn
}

func Dial(URL string) (*Conn, error) {
	c, _, err := websocket.DefaultDialer.Dial(URL, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{c}, nil
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func Upgrade(w http.ResponseWriter, r *http.Request) (*Conn, error) {

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return &Conn{c}, nil
}
