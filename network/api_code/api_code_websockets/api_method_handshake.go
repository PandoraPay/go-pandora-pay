package api_code_websockets

import (
	"pandora-pay/config"
	"pandora-pay/network/network_config"
	"pandora-pay/network/websocks/connection"
)

func Handshake(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return &connection.ConnectionHandshake{config.NAME, config.VERSION_STRING, config.NETWORK_SELECTED, config.NODE_CONSENSUS, network_config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING}, nil
}
