package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	api_store "pandora-pay/network/api/api-store"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
)

type APIWebsockets struct {
	GetMap   map[string]func(conn *connection.AdvancedConnection, values []byte) interface{}
	chain    *blockchain.Blockchain
	mempool  *mempool.Mempool
	apiStore *api_store.APIStore
}

func (api *APIWebsockets) ValidateHandshake(handshake *APIHandshake) error {
	handshake2 := *handshake
	if handshake2[2] != string(config.NETWORK_SELECTED) {
		return errors.New("Network is different")
	}
	return nil
}

func (api *APIWebsockets) handshake(conn *connection.AdvancedConnection, values []byte) interface{} {
	handshake := APIHandshake{}
	if err := json.Unmarshal(values, &handshake); err != nil {
		panic(err)
	}
	if err := api.ValidateHandshake(&handshake); err != nil {
		panic(err)
	}
	return &APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)}
}

func (api *APIWebsockets) hash(conn *connection.AdvancedConnection, values []byte) interface{} {

	blockHeight := APIBlockHeight(0)
	if err := json.Unmarshal(values, &blockHeight); err != nil {
		panic(err)
	}

	return nil
}

func CreateWebsocketsAPI(apiStore *api_store.APIStore, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *APIWebsockets {

	api := APIWebsockets{
		chain:    chain,
		mempool:  mempool,
		apiStore: apiStore,
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) interface{}{
		"handshake": api.handshake,
		"hash":      api.hash,
	}

	return &api
}
