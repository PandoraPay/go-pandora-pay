package api_websockets

import (
	"pandora-pay/network/websocks/connection"
)

type APILogoutReply struct {
	Status bool `json:"status" msgpack:""`
}

func (api *APIWebsockets) GetLogout_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	reply := &APILogoutReply{}

	if conn.Authenticated.IsNotSet() {
		return reply, nil
	}

	conn.Authenticated.UnSet()
	reply.Status = true

	return reply, nil
}
