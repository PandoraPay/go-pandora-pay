package websocks

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
)

type APIWebsockets struct {
	GetMap  map[string]func(values url.Values) interface{}
	chain   *blockchain.Blockchain
	mempool *mempool.Mempool
}

func (api *APIWebsockets) handshake(values url.Values) interface{} {

	//if values.Get("Network") != config.NETWORK_SELECTED {
	//	return errors.New("Network is different")
	//}

	return &struct {
		Name    string
		Version string
		Network uint64
	}{
		Name:    config.NAME,
		Version: config.VERSION,
		Network: config.NETWORK_SELECTED,
	}
}

func CreateWebsocketsAPI(chain *blockchain.Blockchain, mempool *mempool.Mempool) *APIWebsockets {

	api := APIWebsockets{
		chain:   chain,
		mempool: mempool,
	}

	api.GetMap = map[string]func(values url.Values) interface{}{
		"handshake": api.handshake,
	}

	return &api
}
