package config

import (
	"math/big"
	"pandora-pay/globals"
)

var (
	NETWORK_SELECTED uint64 = 0
	DEBUG            bool   = false
	CPU_THREADS      int    = 1
	ARHITECTURE      string = ""
	OS               string = ""
)

var (
	NAME    = "PANDORA PAY"
	VERSION = "0.0"

	MAIN_NET_NETWORK_BYTE        uint64 = 0
	MAIN_NET_NETWORK_BYTE_PREFIX        = "PANDORA" // must have 7 characters

	TEST_NET_NETWORK_BYTE        uint64 = 1033
	TEST_NET_NETWORK_BYTE_PREFIX        = "PANTEST" // must have 7 characters

	DEV_NET_NETWORK_BYTE        uint64 = 4255
	DEV_NET_NETWORK_BYTE_PREFIX        = "PANDDEV" // must have 7 characters

	NETWORK_BYTE_PREFIX_LENGTH = 7

	NETWORK_TIMESTAMP_DRIFT_MAX uint64 = 10

	BLOCK_MAX_SIZE          uint64 = 1 << 10
	BLOCK_TIME              uint64 = 60 //seconds
	DIFFICULTY_BLOCK_WINDOW uint64 = 10

	BIG_INT_ZERO    = big.NewInt(0)
	BIG_INT_ONE     = big.NewInt(1)
	BIG_INT_MAX_256 = new(big.Int).Lsh(BIG_INT_ONE, 256) // 0xFFFFFFFF....

	BIG_FLOAT_MAX_256 = new(big.Float).SetInt(BIG_INT_MAX_256) // 0xFFFFFFFF....
)

func InitConfig() {

	if globals.Arguments["--testnet"] == true {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
	}

	if globals.Arguments["--devnet"] == true {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
	}

	if globals.Arguments["--debug"] == true {
		DEBUG = true
	}

}
