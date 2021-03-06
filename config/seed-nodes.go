package config

import "net/url"

var (
	MAIN_NET_SEED_NODES = []url.URL{
		{Scheme: "ws", Host: "127.0.0.1:5231", Path: "/ws"},
	}

	TEST_NET_SEED_NODES = []url.URL{
		{Scheme: "ws", Host: "helloworldx.ddns.net:16000", Path: "/ws"},
		{Scheme: "ws", Host: "helloworldx.ddns.net:16001", Path: "/ws"},
		{Scheme: "ws", Host: "helloworldx.ddns.net:16002", Path: "/ws"},
		{Scheme: "ws", Host: "helloworldx.ddns.net:16003", Path: "/ws"},
	}

	DEV_NET_SEED_NODES = []url.URL{
		{Scheme: "ws", Host: "127.0.0.1:5230", Path: "/ws"},
		{Scheme: "ws", Host: "127.0.0.1:5231", Path: "/ws"},
		{Scheme: "ws", Host: "127.0.0.1:5232", Path: "/ws"},
		{Scheme: "ws", Host: "127.0.0.1:5233", Path: "/ws"},
		{Scheme: "ws", Host: "127.0.0.1:5234", Path: "/ws"},
	}
)
