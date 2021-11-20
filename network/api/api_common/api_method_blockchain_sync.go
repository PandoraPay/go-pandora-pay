package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) getBlockchainSync() ([]byte, error) {
	return json.Marshal(api.localChainSync.Load().(*blockchain_sync.BlockchainSyncData))
}

func (api *APICommon) GetBlockchainSync_http(values *url.Values) (interface{}, error) {
	return api.getBlockchainSync()
}

func (api *APICommon) GetBlockchainSync_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getBlockchainSync()
}
