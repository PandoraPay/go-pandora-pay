package config_nodes

import (
	"pandora-pay/config/arguments"
	"strconv"
)

type DelegatorNode struct {
	Url  string `json:"url" msgpack:"url" `
	Name string `json:"name" msgpack:"name" `
}

var (
	MAIN_NET_DELEGATOR_NODES = []*DelegatorNode{}

	TEST_NET_DELEGATOR_NODES = []*DelegatorNode{
		{
			"ws://helloworldx.ddns.net:16000/ws",
			"helloworldx 1",
		},
		{
			"ws://helloworldx.ddns.net:16001/ws",
			"helloworldx 2",
		},
		{
			"ws://helloworldx.ddns.net:16002/ws",
			"helloworldx 3",
		},
		{
			"ws://helloworldx.ddns.net:16003/ws",
			"helloworldx 4",
		},
	}

	DEV_NET_DELEGATOR_NODES = []*DelegatorNode{
		{
			"ws://127.0.0.1:5230/ws",
			"127.0.0.1 1",
		},
		{
			"ws://127.0.0.1:5231/ws",
			"127.0.0.1 2",
		},
		{
			"ws://127.0.0.1:5232/ws",
			"127.0.0.1 3",
		},
		{
			"ws://127.0.0.1:5233/ws",
			"127.0.0.1 4",
		},
		{
			"ws://127.0.0.1:5234/ws",
			"127.0.0.1 5",
		},
	}
)

var (
	/* DELEGATES_ALLOWED_ENABLES
	this will enable accepting delegating for other users their delegated stakes
	*/
	DELEGATOR_ENABLED      = false
	DELEGATOR_REQUIRE_AUTH = false
	DELEGATES_MAXIMUM      = 10000
)

func InitConfig() (err error) {

	if arguments.Arguments["--delegates-maximum"] != nil {
		if DELEGATES_MAXIMUM, err = strconv.Atoi(arguments.Arguments["--delegates-maximum"].(string)); err != nil {
			return
		}
	}

	if arguments.Arguments["--delegator-enabled"] == "true" {
		DELEGATOR_ENABLED = true
	}

	if arguments.Arguments["--delegator-require-auth"] == "true" {
		DELEGATOR_REQUIRE_AUTH = true
	}

	return nil
}
