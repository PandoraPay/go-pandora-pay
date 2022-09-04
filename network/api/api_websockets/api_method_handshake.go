package api_websockets

import (
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) handshake(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return &connection.ConnectionHandshake{config.NAME, config.VERSION_STRING, config.NETWORK_SELECTED, config.CONSENSUS, config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING}, nil
}
