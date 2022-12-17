//go:build !js
// +build !js

package websock

import (
	"github.com/gorilla/websocket"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"net/url"
	"pandora-pay/config/arguments"
)

type Conn struct {
	*websocket.Conn
}

func Dial(URL string) (*Conn, error) {

	//tcp proxy
	var dialer *websocket.Dialer
	if arguments.Arguments["--tcp-proxy"] != nil {

		u, err := url.Parse(arguments.Arguments["--tcp-proxy"].(string)) // some not-exist-proxy
		if err != nil {
			return nil, err
		}

		netDialer, err := proxy.FromURL(u, &net.Dialer{})
		if err != nil {
			return nil, err
		}

		dialer = &websocket.Dialer{NetDial: netDialer.Dial}
	} else {
		dialer = websocket.DefaultDialer
	}

	c, _, err := dialer.Dial(URL, nil)
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
