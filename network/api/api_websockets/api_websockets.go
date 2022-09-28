package api_websockets

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/info"
	"pandora-pay/config"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_delegator_node"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/api/api_websockets/consensus"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
	"pandora-pay/txs_validator"
)

type APIWebsockets struct {
	GetMap                    map[string]func(conn *connection.AdvancedConnection, values []byte) (interface{}, error)
	Consensus                 *consensus.Consensus
	chain                     *blockchain.Blockchain
	mempool                   *mempool.Mempool
	settings                  *settings.Settings
	apiCommon                 *api_common.APICommon
	apiStore                  *api_common.APIStore
	SubscriptionNotifications *multicast.MulticastChannel[*api_types.APISubscriptionNotification]
}

func handleAuthenticated[T any, B any](callback func(r *http.Request, args *T, reply *B, authenticated bool) error) func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
		args := new(T)
		if err := msgpack.Unmarshal(values, args); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply, conn.Authenticated.IsSet())
	}
}

func handle[T any, B any](callback func(r *http.Request, args *T, reply *B) error) func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return func(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
		args := new(T)
		if err := msgpack.Unmarshal(values, args); err != nil {
			return nil, err
		}

		reply := new(B)
		return reply, callback(nil, args, reply)
	}
}

func NewWebsocketsAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool, txsValidator *txs_validator.TxsValidator) *APIWebsockets {

	api := &APIWebsockets{
		nil,
		consensus.NewConsensus(chain, mempool, txsValidator),
		chain,
		mempool,
		settings,
		apiCommon,
		apiStore,
		multicast.NewMulticastChannel[*api_types.APISubscriptionNotification](),
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) (interface{}, error){
		"ping":                    handle[struct{}, api_common.APIPingReply](api.apiCommon.GetPing),
		"":                        handle[struct{}, api_common.APIInfoReply](api.apiCommon.GetInfo),
		"chain":                   handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"blockchain":              handle[struct{}, api_common.APIBlockchain](api.apiCommon.GetBlockchain),
		"blockchain/staking-info": handle[api_common.APIStakingInfoRequest, api_common.APIStakingInfoReply](api.apiCommon.GetStakingInfo),
		"blockchain/genesis-info": handle[api_common.APIGenesisInfoRequest, api_common.APIGenesisInfoReply](api.apiCommon.GetGenesisInfo),
		"blockchain/supply":       handle[struct{}, api_common.APISupply](api.apiCommon.GetSupply),
		"blockchain/supply-only":  handle[struct{}, uint64](api.apiCommon.GetSupplyOnly),
		"sync":                    handle[struct{}, blockchain_sync.BlockchainSyncData](api.apiCommon.GetBlockchainSync),
		"block-hash":              handle[api_common.APIBlockHashRequest, api_common.APIBlockHashReply](api.apiCommon.GetBlockHash),
		"block":                   handle[api_common.APIBlockRequest, api_common.APIBlockReply](api.apiCommon.GetBlock),
		"block/exists":            handle[api_common.APIBlockExistsRequest, api_common.APIBlockExistsReply](api.apiCommon.GetBlockExists),
		"block-complete":          handle[api_common.APIBlockCompleteRequest, api_common.APIBlockCompleteReply](api.apiCommon.GetBlockComplete),
		"tx-hash":                 handle[api_common.APITxHashRequest, api_common.APITxHashReply](api.apiCommon.GetTxHash),
		"tx":                      handle[api_common.APITxRequest, api_common.APITxReply](api.apiCommon.GetTx),
		"tx/exists":               handle[api_common.APITxExistsRequest, api_common.APITxExistsReply](api.apiCommon.GetTxExists),
		"tx-raw":                  handle[api_common.APITxRawRequest, api_common.APITxRawReply](api.apiCommon.GetTxRaw),
		"account":                 handle[api_common.APIAccountRequest, api_common.APIAccountReply](api.apiCommon.GetAccount),
		"accounts/count":          handle[api_common.APIAccountsCountRequest, api_common.APIAccountsCountReply](api.apiCommon.GetAccountsCount),
		"asset":                   handle[api_common.APIAssetRequest, api_common.APIAssetReply](api.apiCommon.GetAsset),
		"asset/exists":            handle[api_common.APIAssetRequest, api_common.APIAssetReply](api.apiCommon.GetAsset),
		"mempool":                 handle[api_common.APIMempoolRequest, api_common.APIMempoolReply](api.apiCommon.GetMempool),
		"mempool/tx-exists":       handle[api_common.APIMempoolExistsRequest, api_common.APIMempoolExistsReply](api.apiCommon.GetMempoolExists),
		"mempool/new-tx":          handle[api_common.APIMempoolNewTxRequest, api_common.APIMempoolNewTxReply](api.apiCommon.MempoolNewTx),
		"network/nodes":           handle[struct{}, api_common.APINetworkNodesReply](api.apiCommon.GetNetworkNodes),
		"wallet/get-addresses":    handleAuthenticated[struct{}, api_common.APIWalletGetAccountsReply](api.apiCommon.GetWalletAddresses),
		"wallet/generate-address": handleAuthenticated[api_common.APIWalletGenerateAddressRequest, api_common.APIWalletGenerateAddressReply](api.apiCommon.GetWalletGenerateAddress),
		"wallet/create-address":   handleAuthenticated[api_common.APIWalletCreateAddressRequest, api_common.APIWalletCreateAddressReply](api.apiCommon.GetWalletCreateAddress),
		"wallet/delete-address":   handleAuthenticated[api_common.APIWalletDeleteAddressRequest, api_common.APIWalletDeleteAddressReply](api.apiCommon.GetWalletDeleteAddress),
		"wallet/get-balances":     handleAuthenticated[api_common.APIWalletGetBalanceRequest, api_common.APIWalletGetBalancesReply](api.apiCommon.GetWalletBalances),
		"wallet/decrypt-tx":       handleAuthenticated[api_common.APIWalletDecryptTxRequest, api_common.APIWalletDecryptTxReply](api.apiCommon.GetWalletDecryptTx),
		//below are ONLY websockets API
		"block-miss-txs":    handle[consensus.APIBlockCompleteMissingTxsRequest, consensus.APIBlockCompleteMissingTxsReply](api.Consensus.GetBlockCompleteMissingTxs),
		"handshake":         api.handshake,
		"mempool/new-tx-id": api.apiCommon.MempoolNewTxId,
		"get-chain":         api.Consensus.GetChain,
		"chain-update":      api.Consensus.ChainUpdate,
		"login":             api.login,
		"logout":            api.logout,
		"sub":               api.subscribe,
		"unsub":             api.unsubscribe,
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

	if config.CONSENSUS == config.CONSENSUS_TYPE_WALLET {
		api.GetMap["sub/notify"] = api.subscribedNotificationReceived
	}

	if api.apiCommon.Faucet != nil {
		api.GetMap["faucet/info"] = handle[struct{}, api_faucet.APIFaucetInfo](api.apiCommon.Faucet.GetFaucetInfo)
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = handle[api_faucet.APIFaucetCoinsRequest, api_faucet.APIFaucetCoinsReply](api.apiCommon.Faucet.GetFaucetCoins)
		}
	}

	if api.apiCommon.DelegatorNode != nil {
		api.GetMap["delegator-node/info"] = handle[struct{}, api_delegator_node.ApiDelegatorNodeInfoReply](api.apiCommon.DelegatorNode.GetDelegatorNodeInfo)
		api.GetMap["delegator-node/notify"] = handleAuthenticated[api_delegator_node.ApiDelegatorNodeNotifyRequest, api_delegator_node.ApiDelegatorNodeNotifyReply](api.apiCommon.DelegatorNode.DelegatorNotify)
	}

	return api
}
