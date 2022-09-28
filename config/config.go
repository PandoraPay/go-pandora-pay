package config

import (
	"errors"
	"github.com/blang/semver/v4"
	"math/big"
	"math/rand"
	"pandora-pay/config/config_auth"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/globals"
	"runtime"
	"strconv"
	"time"
)

var (
	DEBUG              = false
	CPU_THREADS        = 1
	ARCHITECTURE       = ""
	OS                 = ""
	NAME               = "WEBDOLLAR"
	VERSION            = semver.MustParse("0.0.1-test.0")
	VERSION_STRING     = VERSION.String()
	BUILD_VERSION      = ""
	LIGHT_COMPUTATIONS = false
	ORIGINAL_PATH      = "" //the original path where the software is located
)

const (
	TRANSACTIONS_MAX_DATA_LENGTH = 512
)

var (
	MAIN_NET_NETWORK_BYTE           uint64 = 0
	MAIN_NET_NETWORK_BYTE_PREFIX           = []byte{88, 64, 67, 254} // must have 7 characters
	MAIN_NET_NETWORK_NAME                  = "MAIN"                  // must have 7 characters
	TEST_NET_NETWORK_BYTE           uint64 = 1033
	TEST_NET_NETWORK_BYTE_PREFIX           = []byte{88, 64, 67, 182} // must have 7 characters
	TEST_NET_NETWORK_NAME                  = "TEST"                  // must have 7 characters
	DEV_NET_NETWORK_BYTE            uint64 = 4255
	DEV_NET_NETWORK_BYTE_PREFIX            = []byte{88, 64, 67, 118} // must have 7 characters
	DEV_NET_NETWORK_NAME                   = "DEV"                   // must have 7 characters
	NETWORK_BYTE_PREFIX_LENGTH             = 4
	NETWORK_TIMESTAMP_DRIFT_MAX     uint64 = 10
	NETWORK_TIMESTAMP_DRIFT_MAX_INT int64  = 10
)

const (
	BLOCK_MAX_SIZE          uint64 = 1024 * 1024
	BLOCK_TIME              uint64 = 30 //seconds
	DIFFICULTY_BLOCK_WINDOW uint64 = 10
	FORK_MAX_UNCLE_ALLOWED  uint64 = 60
	FORK_MAX_DOWNLOAD       uint64 = 20
)

var (
	NETWORK_SELECTED                 = MAIN_NET_NETWORK_BYTE
	NETWORK_SELECTED_BYTE_PREFIX     = MAIN_NET_NETWORK_BYTE_PREFIX
	NETWORK_SELECTED_NAME            = MAIN_NET_NETWORK_NAME
	NETWORK_SELECTED_SEEDS           = MAIN_NET_SEED_NODES
	NETWORK_SELECTED_DELEGATOR_NODES = config_nodes.MAIN_NET_DELEGATOR_NODES
	WEBSOCKETS_NETWORK_CLIENTS_MAX   = int64(50)
	WEBSOCKETS_NETWORK_SERVER_MAX    = int64(500)
)

const (
	WEBSOCKETS_TIMEOUT                            = 5 * time.Second //seconds
	WEBSOCKETS_MAX_READ_THREADS                   = 5
	WEBSOCKETS_PONG_WAIT                          = 60 * time.Second // Time allowed to read the next pong message from the peer.
	WEBSOCKETS_PING_INTERVAL                      = (WEBSOCKETS_PONG_WAIT * 8) / 10
	WEBSOCKETS_MAX_READ                           = BLOCK_MAX_SIZE + 5*1024
	WEBSOCKETS_MAX_SUBSCRIPTIONS                  = 30
	WEBSOCKETS_INCREASE_KNOWN_NODE_SCORE_INTERVAL = 1 * time.Minute
)

var (
	API_MEMPOOL_MAX_TRANSACTIONS = 50
	API_ACCOUNT_MAX_TXS          = uint64(10)
	API_ASSETS_INFO_MAX_RESULTS  = 10
)

var (
	BIG_INT_ZERO      = big.NewInt(0)
	BIG_INT_ONE       = big.NewInt(1)
	BIG_INT_MAX_256   = new(big.Int).Lsh(BIG_INT_ONE, 256)     // 0xFFFFFFFF....
	BIG_FLOAT_MAX_256 = new(big.Float).SetInt(BIG_INT_MAX_256) // 0xFFFFFFFF....
)

var (
	INSTANCE    = ""
	INSTANCE_ID = 0
)

var (
	CONSENSUS              ConsensusType = CONSENSUS_TYPE_FULL
	SEED_WALLET_NODES_INFO bool
)

var (
	NETWORK_ADDRESS_URL_STRING           string
	NETWORK_WEBSOCKET_ADDRESS_URL_STRING string
	NETWORK_KNOWN_NODES_LIMIT            int32 = 5000
	NETWORK_KNOWN_NODES_LIST_RETURN            = 100
)

func InitConfig() (err error) {

	if globals.Arguments["--network"] == "mainnet" {

	} else if globals.Arguments["--network"] == "testnet" {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = TEST_NET_SEED_NODES
		NETWORK_SELECTED_DELEGATOR_NODES = config_nodes.TEST_NET_DELEGATOR_NODES
		NETWORK_SELECTED_NAME = TEST_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = TEST_NET_NETWORK_BYTE_PREFIX
	} else if globals.Arguments["--network"] == "devnet" {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = DEV_NET_SEED_NODES
		NETWORK_SELECTED_DELEGATOR_NODES = config_nodes.DEV_NET_DELEGATOR_NODES
		NETWORK_SELECTED_NAME = DEV_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = DEV_NET_NETWORK_BYTE_PREFIX
	} else {
		return errors.New("selected --network is invalid. Accepted only: mainnet, testnet, devnet")
	}

	if globals.Arguments["--debug"] == true {
		DEBUG = true
	}

	if globals.Arguments["--tcp-max-clients"] != nil {
		if WEBSOCKETS_NETWORK_CLIENTS_MAX, err = strconv.ParseInt(globals.Arguments["--tcp-max-clients"].(string), 10, 64); err != nil {
			return
		}
	}

	if globals.Arguments["--tcp-max-server-sockets"] != nil {
		if WEBSOCKETS_NETWORK_SERVER_MAX, err = strconv.ParseInt(globals.Arguments["--tcp-max-server-sockets"].(string), 10, 64); err != nil {
			return
		}
	}

	SEED_WALLET_NODES_INFO = false
	switch globals.Arguments["--consensus"] {
	case "full":
		CONSENSUS = CONSENSUS_TYPE_FULL
		if globals.Arguments["--seed-wallet-nodes-info"] == "true" {
			SEED_WALLET_NODES_INFO = true
		}
	case "wallet":
		CONSENSUS = CONSENSUS_TYPE_WALLET
	case "none":
		CONSENSUS = CONSENSUS_TYPE_NONE
	default:
		return errors.New("invalid consensus argument")
	}

	if globals.Arguments["--light-computations"] == true {
		LIGHT_COMPUTATIONS = true
	}

	if NETWORK_SELECTED == TEST_NET_NETWORK_BYTE || NETWORK_SELECTED == DEV_NET_NETWORK_BYTE {

		if globals.Arguments["--hcaptcha-secret"] != nil {
			HCAPTCHA_SECRET_KEY = globals.Arguments["--hcaptcha-secret"].(string)
		}

		if HCAPTCHA_SECRET_KEY != "" && globals.Arguments["--faucet-testnet-enabled"] == "true" {
			FAUCET_TESTNET_ENABLED = true
		}

	}

	if err = config_nodes.InitConfig(); err != nil {
		return
	}

	if err = config_auth.InitConfig(); err != nil {
		return
	}

	if err = config_init(); err != nil {
		return
	}

	if err = config_forging.InitConfig(); err != nil {
		return
	}

	return
}

func init() {
	rand.Seed(time.Now().UnixNano())
	CPU_THREADS = runtime.GOMAXPROCS(0)
	ARCHITECTURE = runtime.GOARCH
	OS = runtime.GOOS
}
