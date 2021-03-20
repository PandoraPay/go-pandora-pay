package api

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
)

type APIWebsockets struct {
	GetMap  map[string]func(values []byte) interface{}
	chain   *blockchain.Blockchain
	mempool *mempool.Mempool
}

func (api *APIWebsockets) handshake(values []byte) interface{} {

	handshake := &APIHandshake{}
	if err := json.Unmarshal(values, handshake); err != nil {
		panic(err)
	}

	if handshake.Network != config.NETWORK_SELECTED {
		panic(errors.New("Network is different"))
	}

	return &APIHandshake{
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

	api.GetMap = map[string]func(values []byte) interface{}{
		"handshake": api.handshake,
	}

	return &api
}
