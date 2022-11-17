package api_http

import (
	"io"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/info"
	"pandora-pay/config"
	"pandora-pay/network/api_code/api_code_http"
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/api_implementation/api_common/api_delegator_node"
	"pandora-pay/network/api_implementation/api_common/api_faucet"
)

type API struct {
	GetMap    map[string]func(values url.Values) (interface{}, error)
	PostMap   map[string]func(values io.ReadCloser) (interface{}, error)
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
		"ping":                    api_code_http.Handle[struct{}, api_common.APIPingReply](api.apiCommon.GetPing),
		"":                        api_code_http.Handle[struct{}, api_common.APIInfoReply](api.apiCommon.GetInfo),
		"chain":                   api_code_http.Handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"blockchain":              api_code_http.Handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"blockchain/staking-info": api_code_http.Handle[api_common.APIStakingInfoRequest, api_common.APIStakingInfoReply](api.apiCommon.GetStakingInfo),
		"blockchain/genesis-info": api_code_http.Handle[api_common.APIGenesisInfoRequest, api_common.APIGenesisInfoReply](api.apiCommon.GetGenesisInfo),
		"blockchain/supply":       api_code_http.Handle[struct{}, api_common.APISupply](api.apiCommon.GetSupply),
		"blockchain/supply-only":  api_code_http.Handle[struct{}, uint64](api.apiCommon.GetSupplyOnly),
		"sync":                    api_code_http.Handle[struct{}, blockchain_sync.BlockchainSyncData](api.apiCommon.GetBlockchainSync),
		"block-hash":              api_code_http.Handle[api_common.APIBlockHashRequest, api_common.APIBlockHashReply](api.apiCommon.GetBlockHash),
		"block/exists":            api_code_http.Handle[api_common.APIBlockExistsRequest, api_common.APIBlockExistsReply](api.apiCommon.GetBlockExists),
		"block":                   api_code_http.Handle[api_common.APIBlockRequest, api_common.APIBlockReply](api.apiCommon.GetBlock),
		"block-complete":          api_code_http.Handle[api_common.APIBlockCompleteRequest, api_common.APIBlockCompleteReply](api.apiCommon.GetBlockComplete),
		"tx-hash":                 api_code_http.Handle[api_common.APITxHashRequest, api_common.APITxHashReply](api.apiCommon.GetTxHash),
		"tx":                      api_code_http.Handle[api_common.APITxRequest, api_common.APITxReply](api.apiCommon.GetTx),
		"tx/exists":               api_code_http.Handle[api_common.APITxExistsRequest, api_common.APITxExistsReply](api.apiCommon.GetTxExists),
		"tx-raw":                  api_code_http.Handle[api_common.APITxRawRequest, api_common.APITxRawReply](api.apiCommon.GetTxRaw),
		"account":                 api_code_http.Handle[api_common.APIAccountRequest, api_common.APIAccountReply](api.apiCommon.GetAccount),
		"accounts/count":          api_code_http.Handle[api_common.APIAccountsCountRequest, api_common.APIAccountsCountReply](api.apiCommon.GetAccountsCount),
		"accounts/keys-by-index":  api_code_http.Handle[api_common.APIAccountsKeysByIndexRequest, api_common.APIAccountsKeysByIndexReply](api.apiCommon.GetAccountsKeysByIndex),
		"accounts/by-keys":        api_code_http.Handle[api_common.APIAccountsByKeysRequest, api_common.APIAccountsByKeysReply](api.apiCommon.GetAccountsByKeys),
		"asset":                   api_code_http.Handle[api_common.APIAssetRequest, api_common.APIAssetReply](api.apiCommon.GetAsset),
		"asset/exists":            api_code_http.Handle[api_common.APIAssetExistsRequest, api_common.APIAssetExistsReply](api.apiCommon.GetAssetExists),
		"asset/fee-liquidity":     api_code_http.Handle[api_common.APIAssetFeeLiquidityFeeRequest, api_common.APIAssetFeeLiquidityFeeReply](api.apiCommon.GetAssetFeeLiquidity),
		"mempool":                 api_code_http.Handle[api_common.APIMempoolRequest, api_common.APIMempoolReply](api.apiCommon.GetMempool),
		"mempool/tx-exists":       api_code_http.Handle[api_common.APIMempoolExistsRequest, api_common.APIMempoolExistsReply](api.apiCommon.GetMempoolExists),
		"mempool/new-tx":          api_code_http.Handle[api_common.APIMempoolNewTxRequest, api_common.APIMempoolNewTxReply](api.apiCommon.MempoolNewTx),
		"network/nodes":           api_code_http.Handle[struct{}, api_common.APINetworkNodesReply](api.apiCommon.GetNetworkNodes),
		"wallet/get-addresses":    api_code_http.HandleAuthenticated[struct{}, api_common.APIWalletGetAccountsReply](api.apiCommon.GetWalletAddresses),
		"wallet/generate-address": api_code_http.HandleAuthenticated[api_common.APIWalletGenerateAddressRequest, api_common.APIWalletGenerateAddressReply](api.apiCommon.GetWalletGenerateAddress),
		"wallet/create-address":   api_code_http.HandleAuthenticated[api_common.APIWalletCreateAddressRequest, api_common.APIWalletCreateAddressReply](api.apiCommon.GetWalletCreateAddress),
		"wallet/delete-address":   api_code_http.HandleAuthenticated[api_common.APIWalletDeleteAddressRequest, api_common.APIWalletDeleteAddressReply](api.apiCommon.GetWalletDeleteAddress),
		"wallet/get-balances":     api_code_http.HandleAuthenticated[api_common.APIWalletGetBalanceRequest, api_common.APIWalletGetBalancesReply](api.apiCommon.GetWalletBalances),
		"wallet/decrypt-tx":       api_code_http.HandleAuthenticated[api_common.APIWalletDecryptTxRequest, api_common.APIWalletDecryptTxReply](api.apiCommon.GetWalletDecryptTx),
	}

	api.PostMap = map[string]func(values io.ReadCloser) (interface{}, error){
		"wallet/private-transfer": api_code_http.HandlePOSTAuthenticated[api_common.APIWalletPrivateTransferRequest, api_common.APIWalletPrivateTransferReply](api.apiCommon.WalletPrivateTransfer),
	}

	if config.NODE_PROVIDE_EXTENDED_INFO_APP {
		api.GetMap["asset-info"] = api_code_http.Handle[api_common.APIAssetInfoRequest, info.AssetInfo](api.apiCommon.GetAssetInfo)
		api.GetMap["block-info"] = api_code_http.Handle[api_common.APIBlockInfoRequest, info.BlockInfo](api.apiCommon.GetBlockInfo)
		api.GetMap["tx-info"] = api_code_http.Handle[api_common.APITransactionInfoRequest, info.TxInfo](api.apiCommon.GetTxInfo)
		api.GetMap["tx-preview"] = api_code_http.Handle[api_common.APITransactionPreviewRequest, api_common.APITransactionPreviewReply](api.apiCommon.GetTxPreview)
		api.GetMap["account/txs"] = api_code_http.Handle[api_common.APIAccountTxsRequest, api_common.APIAccountTxsReply](api.apiCommon.GetAccountTxs)
		api.GetMap["account/mempool"] = api_code_http.Handle[api_common.APIAccountMempoolRequest, api_common.APIAccountMempoolReply](api.apiCommon.GetAccountMempool)
		api.GetMap["account/mempool-nonce"] = api_code_http.Handle[api_common.APIAccountMempoolNonceRequest, api_common.APIAccountMempoolNonceReply](api.apiCommon.GetAccountMempoolNonce)
	}

	if api.apiCommon.Faucet != nil {
		api.GetMap["faucet/info"] = api_code_http.Handle[struct{}, api_faucet.APIFaucetInfo](api.apiCommon.Faucet.GetFaucetInfo)
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api_code_http.Handle[api_faucet.APIFaucetCoinsRequest, api_faucet.APIFaucetCoinsReply](api.apiCommon.Faucet.GetFaucetCoins)
		}
	}

	if api.apiCommon.DelegatorNode != nil {
		api.GetMap["delegator-node/info"] = api_code_http.Handle[struct{}, api_delegator_node.ApiDelegatorNodeInfoReply](api.apiCommon.DelegatorNode.GetDelegatorNodeInfo)
		api.GetMap["delegator-node/notify"] = api_code_http.HandleAuthenticated[api_delegator_node.ApiDelegatorNodeNotifyRequest, api_delegator_node.ApiDelegatorNodeNotifyReply](api.apiCommon.DelegatorNode.DelegatorNotify)
	}

	return &api
}
