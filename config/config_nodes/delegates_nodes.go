package config_nodes

import (
	"math"
	"net/url"
	"pandora-pay/config/config_stake"
)

type DelegateNode struct {
	Url  *url.URL `json:"url"`
	Name string   `json:"name"`
}

var (
	MAIN_NET_DELEGATES_NODES = []*DelegateNode{}

	TEST_NET_DELEGATES_NODES = []*DelegateNode{
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16000", Path: "/ws"},
			"helloworldx 1",
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16001", Path: "/ws"},
			"helloworldx 2",
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16002", Path: "/ws"},
			"helloworldx 3",
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16003", Path: "/ws"},
			"helloworldx 4",
		},
	}

	DEV_NET_DELEGATES_NODES = []*DelegateNode{
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5230", Path: "/ws"},
			"127.0.0.1 1",
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5231", Path: "/ws"},
			"127.0.0.1 2",
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5232", Path: "/ws"},
			"127.0.0.1 3",
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5233", Path: "/ws"},
			"127.0.0.1 4",
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5234", Path: "/ws"},
			"127.0.0.1 5",
		},
	}
)

var (
	DELEGATES_ALLOWED_ACTIVATED = false
	DELEGATES_MAXIMUM           = 10000
	DELEGATES_FEE               = uint16(math.Floor(float64(0.00 * config_stake.DELEGATING_STAKING_FEE_MAX_VALUE))) // max DELEGATING_STAKING_FEE_MAX_VALUE
)
