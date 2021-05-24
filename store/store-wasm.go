// +build wasm

package store

import (
	"os"
	"pandora-pay/config"
)

func create_db() (err error) {

	var prefix = ""

	switch config.NETWORK_SELECTED {
	case config.MAIN_NET_NETWORK_BYTE:
		prefix += "main"
	case config.TEST_NET_NETWORK_BYTE:
		prefix += "test"
	case config.DEV_NET_NETWORK_BYTE:
		prefix += "dev"
	default:
		panic("Network is unknown")
	}

	if _, err = os.Stat("./" + prefix); os.IsNotExist(err) {
		if err = os.Mkdir("./"+prefix, 0755); err != nil {
			return
		}
	}

	StoreBlockchain = &Store{Name: prefix + "/blockchain"}
	StoreWallet = &Store{Name: prefix + "/wallet"}
	StoreSettings = &Store{Name: prefix + "/settings"}
	StoreMempool = &Store{Name: prefix + "/mempool"}

	StoreBlockchain.init()
	StoreWallet.init()
	StoreSettings.init()
	StoreMempool.init()

	return
}
