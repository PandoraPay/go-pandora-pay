package config_auth

import (
	"encoding/json"
	"pandora-pay/config/globals"
)

type ConfigAuth struct {
	Username string `json:"user"`
	Password string `json:"pass"`
}

var (
	CONFIG_AUTH_USERS_LIST []*ConfigAuth
	CONFIG_AUTH_USERS_MAP  map[string]*ConfigAuth
)

func InitConfig() (err error) {

	if str := globals.Arguments["--auth-users"]; str != nil {
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
