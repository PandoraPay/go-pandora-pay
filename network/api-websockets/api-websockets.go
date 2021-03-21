package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
)

type APIWebsockets struct {
	GetMap  map[string]func(conn *connection.AdvancedConnection, values []byte) interface{}
	chain   *blockchain.Blockchain
	mempool *mempool.Mempool
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

func CreateWebsocketsAPI(chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *APIWebsockets {

	api := APIWebsockets{
		chain:   chain,
		mempool: mempool,
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) interface{}{
		"handshake": api.handshake,
	}

	return &api
}
