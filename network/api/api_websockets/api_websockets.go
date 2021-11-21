package api_websockets

import (
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
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
		"mempool":                api.apiCommon.GetMempool_websockets,
		"mempool/tx-exists":      api.apiCommon.GetMempoolExists_websockets,
		"mempool/new-tx":         api.apiCommon.MempoolNewTx_websockets,
		"network/nodes-list":     api.apiCommon.GetNetworkNodesList_websockets,
		//below are ONLY websockets API
		"block-miss-txs":    api.apiCommon.GetBlockCompleteMissingTxs_websockets,
		"handshake":         api.getHandshake,
		"mempool/new-tx-id": api.apiCommon.MempoolNewTxId_websockets,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["sub"] = api.subscribe
		api.GetMap["unsub"] = api.unsubscribe
		api.GetMap["asset-info"] = api.apiCommon.GetAssetInfo_websockets
		api.GetMap["block-info"] = api.apiCommon.GetBlockInfo_websockets
		api.GetMap["tx-info"] = api.apiCommon.GetTxInfo_websockets
		api.GetMap["tx-preview"] = api.apiCommon.GetTxPreview_websockets
		api.GetMap["account/txs"] = api.apiCommon.GetAccountTxs_websockets
		api.GetMap["account/mempool"] = api.apiCommon.GetAccountMempool_websockets
		api.GetMap["account/mempool-nonce"] = api.apiCommon.GetAccountMempoolNonce_websockets
	}

	if config.SEED_WALLET_NODES_INFO || config.CONSENSUS == config.CONSENSUS_TYPE_WALLET {
		api.GetMap["sub/notify"] = api.subscribedNotificationReceived
	}

	if api.apiCommon.APICommonFaucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.APICommonFaucet.GetFaucetInfo_websockets
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.APICommonFaucet.GetFaucetCoins_websockets
		}
	}

	if api.apiCommon.APIDelegatorNode != nil {
		api.GetMap["delegator-node/info"] = api.apiCommon.APIDelegatorNode.GetDelegatorNodeInfo_websockets
		api.GetMap["delegator-node/ask"] = api.apiCommon.APIDelegatorNode.GetDelegatorNodeAsk_websockets
	}

	return api
}
