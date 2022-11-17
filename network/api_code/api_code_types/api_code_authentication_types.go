package api_code_types

import (
	"net/url"
	"pandora-pay/network/network_config/network_config_auth"
)

func CheckAuthenticated(args url.Values) bool {

	user := network_config_auth.CONFIG_AUTH_USERS_MAP[args.Get("user")]
	if user == nil {
		return false
	}

	return user.Password == args.Get("pass")
}

type APIAuthenticated[T any] struct {
	User string `json:"user" msgpack:"user"`
	Pass string `json:"pass" msgpack:"pass"`
	Data *T     `json:"req" msgpack:"req"`
}

func (authenticated *APIAuthenticated[T]) CheckAuthenticated() bool {
	user := network_config_auth.CONFIG_AUTH_USERS_MAP[authenticated.User]
	if user == nil {
		return false
	}

	return user.Password == authenticated.Pass
}
