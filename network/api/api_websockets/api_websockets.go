package api_websockets

import (
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
)

type APIWebsockets struct {
	GetMap                    map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain                     *blockchain.Blockchain
	mempool                   *mempool.Mempool
	settings                  *settings.Settings
	apiCommon                 *api_common.APICommon
	apiStore                  *api_common.APIStore
	SubscriptionNotifications *multicast.MulticastChannel //*api_common.APISubscriptionNotification
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return json.Marshal(&connection.ConnectionHandshake{config.NAME, config.VERSION, config.NETWORK_SELECTED, config.CONSENSUS, config.NETWORK_ADDRESS_URL_STRING, nil})
}

func (api *APIWebsockets) getBlockInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *APIWebsockets) getAccountTxs(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountTxsRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountTxs(request)
}

func (api *APIWebsockets) getAccountMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountBaseRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountMempool(request)
}

func (api *APIWebsockets) getAccountMempoolNonce(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountBaseRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountMempoolNonce(request)
}

func (api *APIWebsockets) getAssetInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAssetInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAssetInfo(request)
}

func (api *APIWebsockets) getTxInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITransactionInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTxInfo(request)
}

func (api *APIWebsockets) getTxPreview(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITransactionInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTxPreview(request)
}

func CreateWebsocketsAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *APIWebsockets {

	api := &APIWebsockets{
		chain:                     chain,
		apiStore:                  apiStore,
		apiCommon:                 apiCommon,
		settings:                  settings,
		mempool:                   mempool,
		SubscriptionNotifications: multicast.NewMulticastChannel(),
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error){
		"":                       api.apiCommon.GetInfo_websockets,
		"chain":                  api.apiCommon.GetBlockchain_websockets,
		"blockchain":             api.apiCommon.GetBlockchain_websockets,
		"sync":                   api.apiCommon.GetBlockchainSync_websockets,
		"ping":                   api.apiCommon.GetPing_websockets,
		"block-hash":             api.apiCommon.GetBlockHash_websockets,
		"block":                  api.apiCommon.GetBlock_websockets,
		"block-complete":         api.apiCommon.GetBlockComplete_websockets,
		"tx":                     api.apiCommon.GetTx_websockets,
		"tx-hash":                api.apiCommon.GetTxHash_websockets,
		"account":                api.apiCommon.GetAccount_websockets,
		"accounts/count":         api.apiCommon.GetAccountsCount_websockets,
		"accounts/keys-by-index": api.apiCommon.GetAccountsKeysByIndex_websockets,
		"accounts/by-keys":       api.apiCommon.GetAccountsByKeys_websockets,
		"asset":                  api.apiCommon.GetAsset_websockets,
		"asset/fee-liquidity":    api.apiCommon.GetAssetFeeLiquidity_websockets,
		"mem-pool":               api.apiCommon.GetMempool_websockets,
		"mem-pool/tx-exists":     api.apiCommon.GetMempoolExists_websockets,
		"mem-pool/new-tx":        api.apiCommon.MempoolNewTx_websockets,
		//below are ONLY websockets API
		"block-miss-txs":     api.apiCommon.GetBlockCompleteMissingTxs_websockets,
		"handshake":          api.getHandshake,
		"mem-pool/new-tx-id": api.apiCommon.MempoolNewTxId_websockets,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["sub"] = api.subscribe
		api.GetMap["unsub"] = api.unsubscribe
		api.GetMap["asset-info"] = api.getAssetInfo
		api.GetMap["block-info"] = api.getBlockInfo
		api.GetMap["tx-info"] = api.getTxInfo
		api.GetMap["tx-preview"] = api.getTxPreview
		api.GetMap["account/txs"] = api.getAccountTxs
		api.GetMap["account/mem-pool"] = api.getAccountMempool
		api.GetMap["account/mem-pool-nonce"] = api.getAccountMempoolNonce
	}

	if config.SEED_WALLET_NODES_INFO || config.CONSENSUS == config.CONSENSUS_TYPE_WALLET {
		api.GetMap["sub/notify"] = api.subscribedNotificationReceived
	}

	if api.apiCommon.APICommonFaucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.APICommonFaucet.GetFaucetInfoWebsocket
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.APICommonFaucet.GetFaucetCoinsWebsocket
		}
	}

	if api.apiCommon.APIDelegatesNode != nil {
		api.GetMap["delegates/info"] = api.apiCommon.APIDelegatesNode.GetDelegatesInfoWebsocket
		api.GetMap["delegates/ask"] = api.apiCommon.APIDelegatesNode.GetDelegatesAskWebsocket
	}

	return api
}
