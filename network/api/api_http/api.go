package api_http

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/network/api/api_common"
)

type API struct {
	GetMap    map[string]func(values url.Values) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func NewAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values url.Values) (interface{}, error){
		"ping":                   api.apiCommon.GetPing_http,
		"":                       api.apiCommon.GetInfo_http,
		"chain":                  api.apiCommon.GetBlockchain_http,
		"blockchain":             api.apiCommon.GetBlockchain_http,
		"sync":                   api.apiCommon.GetBlockchainSync_http,
		"block-hash":             api.apiCommon.GetBlockHash_http,
		"block":                  api.apiCommon.GetBlock_http,
		"block-complete":         api.apiCommon.GetBlockComplete_http,
		"tx-hash":                api.apiCommon.GetTxHash_http,
		"tx":                     api.apiCommon.GetTx_http,
		"tx-raw":                 api.apiCommon.GetTxRaw_http,
		"account":                api.apiCommon.GetAccount_http,
		"accounts/count":         api.apiCommon.GetAccountsCount_http,
		"accounts/keys-by-index": api.apiCommon.GetAccountsKeysByIndex_http,
		"accounts/by-keys":       api.apiCommon.GetAccountsByKeys_http,
		"asset":                  api.apiCommon.GetAsset_http,
		"asset/fee-liquidity":    api.apiCommon.GetAssetFeeLiquidity_http,
		"mempool":                api.apiCommon.GetMempool_http,
		"mempool/tx-exists":      api.apiCommon.GetMempoolExists_http,
		"mempool/new-tx":         api.apiCommon.MempoolNewTx_http,
		"network/nodes":          api.apiCommon.GetNetworkNodes_http,
		"wallet/get-addresses":   api.apiCommon.WalletGetAddresses_http,
		"wallet/create-address":  api.apiCommon.WalletCreateAddress_http,
		"wallet/delete-address":  api.apiCommon.WalletDeleteAddress_http,
		"wallet/get-balances":    api.apiCommon.WalletGetBalances_http,
		"wallet/decode-tx":       api.apiCommon.WalletDecodeTx_http,
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

	if api.apiCommon.Faucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.Faucet.GetFaucetInfo_http
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.Faucet.GetFaucetCoins_http
		}
	}

	if api.apiCommon.DelegatorNode != nil {
		api.GetMap["delegator-node/info"] = api.apiCommon.DelegatorNode.GetDelegatorNodeInfo_http
		api.GetMap["delegator-node/ask"] = api.apiCommon.DelegatorNode.GetDelegatorNodeAsk_http
	}

	return &api
}
