package config

import (
	"errors"
	"github.com/blang/semver/v4"
	"math/big"
	"math/rand"
	"pandora-pay/config/arguments"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/config_nodes"
	"runtime"
	"time"
)

var (
	DEBUG              = false
	CPU_THREADS        = 1
	ARCHITECTURE       = ""
	OS                 = ""
	NAME               = "PANDORA PAY"
	VERSION            = semver.MustParse("0.0.1-test.0")
	VERSION_STRING     = VERSION.String()
	BUILD_VERSION      = ""
	LIGHT_COMPUTATIONS = false
	ORIGINAL_PATH      = "" //the original path where the software is located
)

const (
	TRANSACTIONS_MAX_DATA_LENGTH = 512
	TRANSACTIONS_ZETHER_RING_MAX = 256
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
	BLOCK_TIME              uint64 = 90 //seconds
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
	NODE_PROVIDE_EXTENDED_INFO_APP bool
	NODE_CONSENSUS                 NodeConsensusType = NODE_CONSENSUS_TYPE_FULL
)

var (
	INSTANCE    = ""
	INSTANCE_ID = 0
)

func InitConfig() (err error) {

	if arguments.Arguments["--network"] == "mainnet" {

	} else if arguments.Arguments["--network"] == "testnet" {
		NETWORK_SELECTED = TEST_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = TEST_NET_SEED_NODES
		NETWORK_SELECTED_DELEGATOR_NODES = config_nodes.TEST_NET_DELEGATOR_NODES
		NETWORK_SELECTED_NAME = TEST_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = TEST_NET_NETWORK_BYTE_PREFIX
	} else if arguments.Arguments["--network"] == "devnet" {
		NETWORK_SELECTED = DEV_NET_NETWORK_BYTE
		NETWORK_SELECTED_SEEDS = DEV_NET_SEED_NODES
		NETWORK_SELECTED_DELEGATOR_NODES = config_nodes.DEV_NET_DELEGATOR_NODES
		NETWORK_SELECTED_NAME = DEV_NET_NETWORK_NAME
		NETWORK_SELECTED_BYTE_PREFIX = DEV_NET_NETWORK_BYTE_PREFIX
	} else {
		return errors.New("selected --network is invalid. Accepted only: mainnet, testnet, devnet")
	}

	if arguments.Arguments["--debug"] == true {
		DEBUG = true
	}

	if arguments.Arguments["--light-computations"] == true {
		LIGHT_COMPUTATIONS = true
	}

	if NETWORK_SELECTED == TEST_NET_NETWORK_BYTE || NETWORK_SELECTED == DEV_NET_NETWORK_BYTE {

		if arguments.Arguments["--hcaptcha-secret"] != nil {
			HCAPTCHA_SECRET_KEY = arguments.Arguments["--hcaptcha-secret"].(string)
		}

		if HCAPTCHA_SECRET_KEY != "" && arguments.Arguments["--faucet-testnet-enabled"] == "true" {
			FAUCET_TESTNET_ENABLED = true
		}

	}

	NODE_PROVIDE_EXTENDED_INFO_APP = false
	switch arguments.Arguments["--node-consensus"] {
	case "full":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_FULL
		if arguments.Arguments["--node-provide-extended-info-app"] == "true" {
			NODE_PROVIDE_EXTENDED_INFO_APP = true
		}
	case "app":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_APP
	case "none":
		NODE_CONSENSUS = NODE_CONSENSUS_TYPE_NONE
	default:
		return errors.New("invalid consensus argument")
	}

	if err = config_nodes.InitConfig(); err != nil {
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
