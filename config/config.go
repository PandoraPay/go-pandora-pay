package config

import "pandora-pay/globals"

var (
	NETWORK_SELECTED uint64 = 0
	DEBUG            bool   = false
	CPU_THREADS      int    = 1
	ARHITECTURE      string = ""
	OS               string = ""
)

const (
	NAME    string = "PANDORA PAY"
	VERSION string = "0.0"

	MAIN_NET_NETWORK_BYTE        uint64 = 0
	MAIN_NET_NETWORK_BYTE_PREFIX string = "PANDORA" // must have 7 characters

	TEST_NET_NETWORK_BYTE        uint64 = 1033
	TEST_NET_NETWORK_BYTE_PREFIX string = "PANTEST" // must have 7 characters

	DEV_NET_NETWORK_BYTE        uint64 = 4255
	DEV_NET_NETWORK_BYTE_PREFIX string = "PANDDEV" // must have 7 characters

	NETWORK_BYTE_PREFIX_LENGTH = 7

	NETWORK_TIMESTAMP_DRIFT_MAX = 10

	BLOCK_MAX_SIZE uint64 = 1 << 10
)

func InitConfig() {

	if globals.Arguments["--testnet"] != nil {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
	}

	if globals.Arguments["--devnet"] != nil {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
	}

	if globals.Arguments["--debug"] == true {
		DEBUG = true
	}

}
