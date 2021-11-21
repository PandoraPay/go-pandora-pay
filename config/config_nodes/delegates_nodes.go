package config_nodes

import (
	"math"
	"pandora-pay/config/config_stake"
)

type DelegateNode struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

var (
	MAIN_NET_DELEGATES_NODES = []*DelegateNode{}

	TEST_NET_DELEGATES_NODES = []*DelegateNode{
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

	DEV_NET_DELEGATES_NODES = []*DelegateNode{
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
	DELEGATES_ALLOWED_ENABLED             = false
	DELEGATES_MAXIMUM                     = 10000
	DELEGATOR_FEE                         = uint64(math.Floor(0.00 * float64(config_stake.DELEGATING_STAKING_FEE_MAX_VALUE))) // max DELEGATING_STAKING_FEE_MAX_VALUE
	DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY = []byte{}
)
