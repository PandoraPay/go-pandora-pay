package api_websockets

import (
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/network/network_config/network_config_auth"
	"pandora-pay/network/websocks/connection"
)

type APILogin struct {
	Username string `json:"user" msgpack:"user"`
	Password string `json:"pass" msgpack:"pass"`
}

type APILoginReply struct {
	Status bool `json:"status" msgpack:"status"`
}

func (api *APIWebsockets) login(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APILogin{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APILoginReply{}

	user := network_config_auth.CONFIG_AUTH_USERS_MAP[args.Username]
	if user == nil || user.Password != args.Password {
		return reply, nil
	}

	conn.Authenticated.Set()
	reply.Status = true

	return reply, nil
}
