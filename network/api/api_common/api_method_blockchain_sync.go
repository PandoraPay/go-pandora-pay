package api_common

import (
	"net/http"
	"net/url"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) BlockchainSync(r *http.Request, args *struct{}, reply *blockchain_sync.BlockchainSyncData) error {
	*reply = *api.localChainSync.Load().(*blockchain_sync.BlockchainSyncData)
	return nil
}

func (api *APICommon) GetBlockchainSync_http(values url.Values) (interface{}, error) {
	var reply *blockchain_sync.BlockchainSyncData
	return reply, api.BlockchainSync(nil, nil, reply)
}

func (api *APICommon) GetBlockchainSync_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	var reply *blockchain_sync.BlockchainSyncData
	return reply, api.BlockchainSync(nil, nil, reply)
}
