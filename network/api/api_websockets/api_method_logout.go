package api_websockets

import (
	"pandora-pay/network/websocks/connection"
)

type APILogoutAnswer struct {
	Status bool `json:"status"`
}

func (api *APIWebsockets) GetLogout_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	reply := &APILogoutAnswer{}

	if !conn.Authenticated {
		return reply, nil
	}

	conn.Authenticated = false
	reply.Status = true

	return reply, nil
}
