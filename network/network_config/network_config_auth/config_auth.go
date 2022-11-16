package network_config_auth

import (
	"encoding/json"
	"pandora-pay/config/arguments"
)

type ConfigAuth struct {
	Username string `json:"user" msgpack:"user"`
	Password string `json:"pass"  msgpack:"pass"`
}

var (
	CONFIG_AUTH_USERS_LIST []*ConfigAuth
	CONFIG_AUTH_USERS_MAP  map[string]*ConfigAuth
)

func InitConfig() (err error) {

	if str := arguments.Arguments["--auth-users"]; str != nil {
		if err = json.Unmarshal([]byte(str.(string)), &CONFIG_AUTH_USERS_LIST); err != nil {
			return
		}
	}

	CONFIG_AUTH_USERS_MAP = map[string]*ConfigAuth{}
	for _, auth := range CONFIG_AUTH_USERS_LIST {
		CONFIG_AUTH_USERS_MAP[auth.Username] = auth
	}

	return
}
