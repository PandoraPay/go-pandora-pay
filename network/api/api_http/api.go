package api_http

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/info"
	"pandora-pay/config"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_delegator_node"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
)

type API struct {
	GetMap    map[string]func(values url.Values) (interface{}, error)
	PostMap   map[string]func(values io.ReadCloser) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func handleAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(values url.Values) (interface{}, error) {
	return func(values url.Values) (interface{}, error) {
		args := new(T)
		if err := urldecoder.Decoder.Decode(args, values); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply, api_types.CheckAuthenticated(values))
	}
}

func handle[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(values url.Values) (interface{}, error) {
	return func(values url.Values) (interface{}, error) {
		args := new(T)
		if err := urldecoder.Decoder.Decode(args, values); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}

func handlePOSTAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(values io.ReadCloser) (interface{}, error) {
	return func(values io.ReadCloser) (interface{}, error) {
		args := new(T)

		decoder := json.NewDecoder(values)
		if err := decoder.Decode(args); err != nil {
			return nil, err
		}

		authenticated := &api_types.APIAuthenticated{}
		if err := decoder.Decode(authenticated); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply, authenticated.CheckAuthenticated())
	}
}

func handlePOST[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(values io.ReadCloser) (interface{}, error) {
	return func(values io.ReadCloser) (interface{}, error) {
		args := new(T)

		if err := json.NewDecoder(values).Decode(args); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}

func NewAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values url.Values) (interface{}, error){
		"ping":                   handle[struct{}, api_common.APIPingReply](api.apiCommon.GetPing),
		"":                       handle[struct{}, api_common.APIInfoReply](api.apiCommon.GetInfo),
		"chain":                  handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"blockchain":             handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"sync":                   handle[struct{}, blockchain_sync.BlockchainSyncData](api.apiCommon.GetBlockchainSync),
		"block-hash":             handle[api_common.APIBlockHashRequest, api_common.APIBlockHashReply](api.apiCommon.GetBlockHash),
		"block":                  handle[api_common.APIBlockRequest, api_common.APIBlockReply](api.apiCommon.GetBlock),
		"block-complete":         handle[api_common.APIBlockCompleteRequest, api_common.APIBlockCompleteReply](api.apiCommon.GetBlockComplete),
		"tx-hash":                handle[api_common.APITxHashRequest, api_common.APITxHashReply](api.apiCommon.GetTxHash),
		"tx":                     handle[api_common.APITransactionRequest, api_common.APITransactionReply](api.apiCommon.GetTx),
		"tx-raw":                 handle[api_common.APITransactionRawRequest, api_common.APITransactionRawReply](api.apiCommon.GetTxRaw),
		"account":                handle[api_common.APIAccountRequest, api_common.APIAccountReply](api.apiCommon.GetAccount),
		"accounts/count":         handle[api_common.APIAccountsCountRequest, api_common.APIAccountsCountReply](api.apiCommon.GetAccountsCount),
		"accounts/keys-by-index": handle[api_common.APIAccountsKeysByIndexRequest, api_common.APIAccountsKeysByIndexReply](api.apiCommon.GetAccountsKeysByIndex),
		"accounts/by-keys":       handle[api_common.APIAccountsByKeysRequest, api_common.APIAccountsByKeysReply](api.apiCommon.GetAccountsByKeys),
		"asset":                  handle[api_common.APIAssetRequest, api_common.APIAssetReply](api.apiCommon.GetAsset),
		"asset/fee-liquidity":    handle[api_common.APIAssetFeeLiquidityFeeRequest, api_common.APIAssetFeeLiquidityFeeReply](api.apiCommon.GetAssetFeeLiquidity),
		"mempool":                handle[api_common.APIMempoolRequest, api_common.APIMempoolReply](api.apiCommon.GetMempool),
		"mempool/tx-exists":      handle[api_common.APIMempoolExistsRequest, api_common.APIMempoolExistsReply](api.apiCommon.GetMempoolExists),
		"mempool/new-tx":         handle[api_common.APIMempoolNewTxRequest, api_common.APIMempoolNewTxReply](api.apiCommon.MempoolNewTx),
		"network/nodes":          handle[struct{}, api_common.APINetworkNodesReply](api.apiCommon.GetNetworkNodes),
		"wallet/get-addresses":   handleAuthenticated[struct{}, api_common.APIWalletGetAccountsReply](api.apiCommon.GetWalletAddresses),
		"wallet/create-address":  handleAuthenticated[api_common.APIWalletCreateAddressRequest, api_common.APIWalletCreateAddressReply](api.apiCommon.GetWalletCreateAddress),
		"wallet/delete-address":  handleAuthenticated[api_common.APIWalletDeleteAddressRequest, api_common.APIWalletDeleteAddressReply](api.apiCommon.GetWalletDeleteAddress),
		"wallet/get-balances":    handleAuthenticated[api_common.APIWalletGetBalanceRequest, api_common.APIWalletGetBalancesReply](api.apiCommon.GetWalletBalances),
		"wallet/decrypt-tx":      handleAuthenticated[api_common.APIWalletDecryptTxRequest, api_common.APIWalletDecryptTxReply](api.apiCommon.GetWalletDecryptTx),
	}

	api.PostMap = map[string]func(values io.ReadCloser) (interface{}, error){
		"wallet/private-transfer": handlePOSTAuthenticated[api_common.APIWalletPrivateTransferRequest, api_common.APIWalletPrivateTransferReply](api.apiCommon.WalletPrivateTransfer),
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["asset-info"] = handle[api_common.APIAssetInfoRequest, info.AssetInfo](api.apiCommon.GetAssetInfo)
		api.GetMap["block-info"] = handle[api_common.APIBlockInfoRequest, info.BlockInfo](api.apiCommon.GetBlockInfo)
		api.GetMap["tx-info"] = handle[api_common.APITransactionInfoRequest, info.TxInfo](api.apiCommon.GetTxInfo)
		api.GetMap["tx-preview"] = handle[api_common.APITransactionPreviewRequest, api_common.APITransactionPreviewReply](api.apiCommon.GetTxPreview)
		api.GetMap["account/txs"] = handle[api_common.APIAccountTxsRequest, api_common.APIAccountTxsReply](api.apiCommon.GetAccountTxs)
		api.GetMap["account/mempool"] = handle[api_common.APIAccountMempoolRequest, api_common.APIAccountMempoolReply](api.apiCommon.GetAccountMempool)
		api.GetMap["account/mempool-nonce"] = handle[api_common.APIAccountMempoolNonceRequest, api_common.APIAccountMempoolNonceReply](api.apiCommon.GetAccountMempoolNonce)
	}

	if api.apiCommon.Faucet != nil {
		api.GetMap["faucet/info"] = handle[struct{}, api_faucet.APIFaucetInfo](api.apiCommon.Faucet.GetFaucetInfo)
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = handle[api_faucet.APIFaucetCoinsRequest, []byte](api.apiCommon.Faucet.GetFaucetCoins)
		}
	}

	if api.apiCommon.DelegatorNode != nil {
		api.GetMap["delegator-node/info"] = handle[struct{}, api_delegator_node.ApiDelegatorNodeInfoReply](api.apiCommon.DelegatorNode.GetDelegatorNodeInfo)
		api.GetMap["delegator-node/ask"] = handle[api_delegator_node.ApiDelegatorNodeAskRequest, api_delegator_node.ApiDelegatorNodeAskReply](api.apiCommon.DelegatorNode.GetDelegatesAsk)
	}

	return &api
}
