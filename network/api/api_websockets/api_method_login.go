package api_websockets

import (
	"encoding/json"
	"pandora-pay/config/config_auth"
	"pandora-pay/network/websocks/connection"
)

type APILogin struct {
	Username string `json:"user"`
	Password string `json:"password"`
}

type APILoginReply struct {
	Status bool `json:"status"`
}

func (api *APIWebsockets) GetLogin_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APILogin{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APILoginReply{}

	user := config_auth.CONFIG_AUTH_USERS_MAP[args.Username]
	if user == nil || user.Password != args.Password {
		return reply, nil
	}

	conn.Authenticated.Set()
	reply.Status = true

	return reply, nil
}
