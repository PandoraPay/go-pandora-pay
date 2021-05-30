package config

import (
	"math/big"
	"math/rand"
	"pandora-pay/config/globals"
	"runtime"
	"time"
)

var (
	DEBUG        = false
	CPU_THREADS  = 1
	ARCHITECTURE = ""
	OS           = ""
	NAME         = "PANDORA PAY"
	VERSION      = "0.01"
)

const (
	MAIN_NET_NETWORK_BYTE           uint64 = 0
	MAIN_NET_NETWORK_BYTE_PREFIX           = "PANDORA" // must have 7 characters
	MAIN_NET_NETWORK_NAME                  = "MAIN"    // must have 7 characters
	TEST_NET_NETWORK_BYTE           uint64 = 1033
	TEST_NET_NETWORK_BYTE_PREFIX           = "PANTEST" // must have 7 characters
	TEST_NET_NETWORK_NAME                  = "TEST"    // must have 7 characters
	DEV_NET_NETWORK_BYTE            uint64 = 4255
	DEV_NET_NETWORK_BYTE_PREFIX            = "PANDDEV" // must have 7 characters
	DEV_NET_NETWORK_NAME                   = "DEV"     // must have 7 characters
	NETWORK_BYTE_PREFIX_LENGTH             = 7
	NETWORK_TIMESTAMP_DRIFT_MAX     uint64 = 10
	NETWORK_TIMESTAMP_DRIFT_MAX_INT int64  = 10
)

const (
	BLOCK_MAX_SIZE          uint64 = 1024 * 1024
	BLOCK_TIME              uint64 = 60 //seconds
	DIFFICULTY_BLOCK_WINDOW uint64 = 10
	FORK_MAX_UNCLE_ALLOWED  uint64 = 60
	FORK_MAX_DOWNLOAD       uint64 = 20
)

var (
	NETWORK_SELECTED             = MAIN_NET_NETWORK_BYTE
	NETWORK_SELECTED_BYTE_PREFIX = MAIN_NET_NETWORK_BYTE_PREFIX
	NETWORK_SELECTED_NAME        = MAIN_NET_NETWORK_NAME
	NETWORK_SELECTED_SEEDS       = MAIN_NET_SEED_NODES
)

const (
	WEBSOCKETS_NETWORK_CLIENTS_MAX = 20
	WEBSOCKETS_NETWORK_SERVER_MAX  = 500
	WEBSOCKETS_TIMEOUT             = 5 * time.Second  //seconds
	WEBSOCKETS_PONG_WAIT           = 60 * time.Second // Time allowed to read the next pong message from the peer.
	WEBSOCKETS_PING_INTERVAL       = (WEBSOCKETS_PONG_WAIT * 8) / 10
	WEBSOCKETS_MAX_READ            = BLOCK_MAX_SIZE + 1024
)

const (
	MEMPOOL_ERROR_BLOCKS = 20
)

var (
	BIG_INT_ZERO      = big.NewInt(0)
	BIG_INT_ONE       = big.NewInt(1)
	BIG_INT_MAX_256   = new(big.Int).Lsh(BIG_INT_ONE, 256)     // 0xFFFFFFFF....
	BIG_FLOAT_MAX_256 = new(big.Float).SetInt(BIG_INT_MAX_256) // 0xFFFFFFFF....
)

var (
	INSTANCE        = ""
	INSTANCE_NUMBER = 0
)

var (
	CONSENSUS ConsensusType = CONSENSUS_TYPE_FULL
)

func StartConfig() {
	rand.Seed(time.Now().UnixNano())
	CPU_THREADS = runtime.GOMAXPROCS(0)
	ARCHITECTURE = runtime.GOARCH
	OS = runtime.GOOS
}

func InitConfig() (err error) {

	if globals.Arguments["--testnet"] == true {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = TEST_NET_SEED_NODES
		NETWORK_SELECTED_NAME = TEST_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = TEST_NET_NETWORK_BYTE_PREFIX
	}

	if globals.Arguments["--devnet"] == true {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = DEV_NET_SEED_NODES
		NETWORK_SELECTED_NAME = DEV_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = DEV_NET_NETWORK_BYTE_PREFIX
	}

	if globals.Arguments["--debug"] == true {
		DEBUG = true
	}

	switch globals.Arguments["--consensus"] {
	case "full":
		CONSENSUS = CONSENSUS_TYPE_FULL
	case "wallet":
		CONSENSUS = CONSENSUS_TYPE_WALLET
	case "none":
		CONSENSUS = CONSENSUS_TYPE_NONE
	default:
		panic("invalid consensus argument")
	}

	if err = config_init(); err != nil {
		return
	}

	return
}
