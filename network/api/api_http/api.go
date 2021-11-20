package api_http

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/network/api/api_common"
)

type API struct {
	GetMap    map[string]func(values *url.Values) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func CreateAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values *url.Values) (interface{}, error){
		"":                       api.apiCommon.GetInfo_http,
		"chain":                  api.apiCommon.GetBlockchain_http,
		"blockchain":             api.apiCommon.GetBlockchain_http,
		"sync":                   api.apiCommon.GetBlockchainSync_http,
		"ping":                   api.apiCommon.GetPing_http,
		"block":                  api.apiCommon.GetBlock_http,
		"block-hash":             api.apiCommon.GetBlockHash_http,
		"block-complete":         api.apiCommon.GetBlockComplete_http,
		"tx":                     api.apiCommon.GetTx_http,
		"tx-hash":                api.apiCommon.GetTxHash_http,
		"account":                api.apiCommon.GetAccount_http,
		"accounts/count":         api.apiCommon.GetAccountsCount_http,
		"accounts/keys-by-index": api.apiCommon.GetAccountsKeysByIndex_http,
		"accounts/by-keys":       api.apiCommon.GetAccountsByKeys_http,
		"asset":                  api.apiCommon.GetAsset_http,
		"asset/fee-liquidity":    api.apiCommon.GetAssetFeeLiquidity_http,
		"mempool":                api.apiCommon.GetMempool_http,
		"mempool/tx-exists":      api.apiCommon.GetMempoolExists_http,
		"mempool/new-tx":         api.apiCommon.MempoolNewTx_http,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["asset-info"] = api.apiCommon.GetAssetInfo_http
		api.GetMap["block-info"] = api.apiCommon.GetBlockInfo_http
		api.GetMap["tx-info"] = api.apiCommon.GetTxInfo_http
		api.GetMap["tx-preview"] = api.apiCommon.GetTxPreview_http
		api.GetMap["account/txs"] = api.apiCommon.GetAccountTxs_http
		api.GetMap["account/mempool"] = api.apiCommon.GetAccountMempool_http
		api.GetMap["account/mempool-nonce"] = api.apiCommon.GetAccountMempoolNonce_http
	}

	if api.apiCommon.APICommonFaucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.APICommonFaucet.GetFaucetInfoHttp
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.APICommonFaucet.GetFaucetCoinsHttp
		}
	}

	if api.apiCommon.APIDelegatesNode != nil {
		api.GetMap["delegates/info"] = api.apiCommon.APIDelegatesNode.GetDelegatesInfoHttp
		api.GetMap["delegates/ask"] = api.apiCommon.APIDelegatesNode.GetDelegatesAskHttp
	}

	return &api
}
