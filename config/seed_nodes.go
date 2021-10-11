package config

import "net/url"

type SeedNode struct {
	Url *url.URL `json:"url"`
}

var (
	MAIN_NET_SEED_NODES = []*SeedNode{}

	TEST_NET_SEED_NODES = []*SeedNode{
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16000", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16001", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16002", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "helloworldx.ddns.net:16003", Path: "/ws"},
		},
	}

	DEV_NET_SEED_NODES = []*SeedNode{
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5230", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5231", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5232", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5233", Path: "/ws"},
		},
		{
			&url.URL{Scheme: "ws", Host: "127.0.0.1:5234", Path: "/ws"},
		},
	}
)
