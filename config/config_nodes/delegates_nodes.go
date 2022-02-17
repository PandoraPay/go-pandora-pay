package config_nodes

import (
	"errors"
	"math"
	"pandora-pay/config/config_stake"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"strconv"
)

type DelegateNode struct {
	Url  string `json:"url" msgpack:"url" `
	Name string `json:"name" msgpack:"name" `
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
	/* DELEGATES_ALLOWED_ENABLES
	this will enable accepting delegating for other users their delegated stakes
	*/
	DELEGATOR_ENABLED                     = false
	DELEGATOR_REQUIRE_AUTH                = false
	DELEGATES_MAXIMUM                     = 10000
	DELEGATOR_FEE                         = uint64(math.Floor(0.00 * float64(config_stake.DELEGATING_STAKING_FEE_MAX_VALUE))) // max DELEGATING_STAKING_FEE_MAX_VALUE
	DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY = []byte{}
	/* DELEGATOR_ACCEPT_CUSTOM_KEYS
	This will allow the full node to accept user generated delegated stakes.
	Be aware that the user can share the private key with other delegators.
	This option should be used only if the reward is NOT divided (like a staking pool)!
	*/
	DELEGATOR_ACCEPT_CUSTOM_KEYS = false
)

func InitConfig() (err error) {

	if globals.Arguments["--delegates-maximum"] != nil {
		if DELEGATES_MAXIMUM, err = strconv.Atoi(globals.Arguments["--delegates-maximum"].(string)); err != nil {
			return
		}
	}

	if globals.Arguments["--delegator-fee"] != nil {
		if DELEGATOR_FEE, err = strconv.ParseUint(globals.Arguments["--delegator-fee"].(string), 10, 64); err != nil {
			return
		}
	}

	if DELEGATOR_FEE > config_stake.DELEGATING_STAKING_FEE_MAX_VALUE {
		return errors.New("DELEGATOR_FEE is invalid")
	}

	if globals.Arguments["--delegator-reward-collector-pub-key"] != nil {
		DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY = helpers.DecodeHex(globals.Arguments["--delegator-reward-collector-pub-key"].(string))
	}

	if globals.Arguments["--delegator-enabled"] == "true" {
		DELEGATOR_ENABLED = true

		if DELEGATOR_FEE > 0 && len(DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY) != cryptography.PublicKeySize {
			return errors.New("DELEGATOR_REWARD_COLLECTOR_PUBLIC_KEY is invalid")
		}
	}

	if globals.Arguments["--delegator-require-auth"] == "true" {
		DELEGATOR_REQUIRE_AUTH = true
	}

	if globals.Arguments["--delegator-accept-custom-keys"] == "true" {
		DELEGATOR_ACCEPT_CUSTOM_KEYS = true
	}

	return nil
}
