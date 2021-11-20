package api_http

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"strconv"
)

type API struct {
	GetMap    map[string]func(values *url.Values) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func (api *API) getBlockInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *API) getAssetInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APIAssetInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAssetInfo(request)
}

func (api *API) getTxInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetTxInfo(request)
}

func (api *API) getTxPreview(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetTxPreview(request)
}

func (api *API) getAccountTxs(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountTxsRequest{}

	var err error
	if values.Get("next") != "" {
		if request.Next, err = strconv.ParseUint(values.Get("next"), 10, 64); err != nil {
			return nil, err
		}
	}

	if err = request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountTxs(request)
}

func (api *API) getAccountMempool(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountBaseRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountMempool(request)
}

func (api *API) getAccountMempoolNonce(values *url.Values) (interface{}, error) {
	request := &api_types.APIAccountBaseRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountMempoolNonce(request)
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
		"mem-pool":               api.apiCommon.GetMempool_http,
		"mem-pool/tx-exists":     api.apiCommon.GetMempoolExists_http,
		"mem-pool/new-tx":        api.apiCommon.MempoolNewTx_http,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["asset-info"] = api.getAssetInfo
		api.GetMap["block-info"] = api.getBlockInfo
		api.GetMap["tx-info"] = api.getTxInfo
		api.GetMap["tx-preview"] = api.getTxPreview
		api.GetMap["account/txs"] = api.getAccountTxs
		api.GetMap["account/mem-pool"] = api.getAccountMempool
		api.GetMap["account/mem-pool-nonce"] = api.getAccountMempoolNonce
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
