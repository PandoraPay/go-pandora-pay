package api_common

import (
	"net/http"
	"pandora-pay/blockchain/blockchain_sync"
)

func (api *APICommon) GetBlockchainSync(r *http.Request, args *struct{}, reply *blockchain_sync.BlockchainSyncData) error {
	*reply = *api.localChainSync.Load()
	return nil
}
