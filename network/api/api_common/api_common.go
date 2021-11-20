package api_common

import (
	"encoding/hex"
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/info"
	"pandora-pay/config"
	"pandora-pay/config/config_nodes"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common/api_delegates_node"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/recovery"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
)

type APICommon struct {
	mempool                   *mempool.Mempool
	chain                     *blockchain.Blockchain
	localChain                *atomic.Value //*APIBlockchain
	localChainSync            *atomic.Value //*blockchain_sync.BlockchainSyncData
	APICommonFaucet           *api_faucet.APICommonFaucet
	APIDelegatesNode          *api_delegates_node.APIDelegatesNode
	ApiStore                  *APIStore
	MempoolDownloadPending    *sync.Map     //[string]chan error
	MempoolProcessedThisBlock *atomic.Value // *sync.Map //[string]error
}

func (api *APICommon) GetBlockInfo(request *api_types.APIBlockInfoRequest) ([]byte, error) {
	blockInfo, err := api.ApiStore.openLoadBlockInfo(request.Height, request.Hash)
	if err != nil || blockInfo == nil {
		return nil, err
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetAccountTxs(request *api_types.APIAccountTxsRequest) ([]byte, error) {

	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	answer, err := api.ApiStore.openLoadAccountTxsFromPublicKey(publicKey, request.Next)
	if err != nil || answer == nil {
		return nil, err
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetAccountMempool(request *api_types.APIAccountBaseRequest) ([]byte, error) {

	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKey)

	var answer []helpers.HexBytes
	if txs != nil {
		answer = make([]helpers.HexBytes, len(txs))
		c := 0
		for _, tx := range txs {
			answer[c] = tx.Tx.Bloom.Hash
			c += 1
		}
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetAccountMempoolNonce(request *api_types.APIAccountBaseRequest) ([]byte, error) {
	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	nonce, err := api.ApiStore.OpenLoadPlainAccountNonceFromPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return json.Marshal(api.mempool.GetNonce(publicKey, nonce))
}

func (api *APICommon) GetTxPreview(request *api_types.APITransactionInfoRequest) ([]byte, error) {
	var txPreview *info.TxPreview
	var txInfo *info.TxInfo
	var err error

	mempool := false
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		txMemPool := api.mempool.Txs.Get(string(request.Hash))
		if txMemPool != nil {
			mempool = true
			if txPreview, err = info.CreateTxPreviewFromTx(txMemPool.Tx); err != nil {
				return nil, err
			}
		} else {
			txPreview, txInfo, err = api.ApiStore.openLoadTxPreview(request.Hash, 0)
		}
	} else {
		txPreview, txInfo, err = api.ApiStore.openLoadTxPreview(nil, request.Height)
	}

	if err != nil || txPreview == nil {
		return nil, err
	}

	result := &api_types.APITransactionPreview{txPreview, mempool, txInfo}
	return json.Marshal(result)
}

func (api *APICommon) GetTxInfo(request *api_types.APITransactionInfoRequest) ([]byte, error) {
	txInfo, err := api.ApiStore.openLoadTxInfo(request.Hash, request.Height)
	if err != nil || txInfo == nil {
		return nil, err
	}
	return json.Marshal(txInfo)
}

func (api *APICommon) GetAssetInfo(request *api_types.APIAssetInfoRequest) ([]byte, error) {
	astInfo, err := api.ApiStore.openLoadAssetInfo(request.Hash, request.Height)
	if err != nil || astInfo == nil {
		return nil, err
	}
	return json.Marshal(astInfo)
}

func (api *APICommon) GetAssetFeeLiquidity(request *APIAssetFeeLiquidityFeeRequest) ([]byte, error) {
	out, err := api.ApiStore.openLoadAssetFeeLiquidity(request.Hash, request.Height)
	if err != nil || out == nil {
		return nil, err
	}
	return json.Marshal(out)
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchain(newChainDataUpdate *blockchain.BlockchainDataUpdate) {
	newLocalChain := &APIBlockchain{
		newChainDataUpdate.Update.Height,
		hex.EncodeToString(newChainDataUpdate.Update.Hash),
		hex.EncodeToString(newChainDataUpdate.Update.PrevHash),
		hex.EncodeToString(newChainDataUpdate.Update.KernelHash),
		hex.EncodeToString(newChainDataUpdate.Update.PrevKernelHash),
		newChainDataUpdate.Update.Timestamp,
		newChainDataUpdate.Update.TransactionsCount,
		newChainDataUpdate.Update.AccountsCount,
		newChainDataUpdate.Update.AssetsCount,
		newChainDataUpdate.Update.Target.String(),
		newChainDataUpdate.Update.BigTotalDifficulty.String(),
	}
	api.localChain.Store(newLocalChain)
	api.MempoolProcessedThisBlock.Store(&sync.Map{})
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchainSync(newLocalSync *blockchain_sync.BlockchainSyncData) {
	api.localChainSync.Store(newLocalSync)
}

func CreateAPICommon(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder, apiStore *APIStore) (api *APICommon, err error) {

	var apiCommonFaucet *api_faucet.APICommonFaucet
	if config.NETWORK_SELECTED == config.TEST_NET_NETWORK_BYTE || config.NETWORK_SELECTED == config.DEV_NET_NETWORK_BYTE {
		if apiCommonFaucet, err = api_faucet.CreateAPICommonFaucet(mempool, chain, wallet, transactionsBuilder); err != nil {
			return
		}
	}

	var apiDelegatesNode *api_delegates_node.APIDelegatesNode
	if config_nodes.DELEGATES_ALLOWED_ENABLED {
		apiDelegatesNode = api_delegates_node.CreateDelegatesNode(chain, wallet)
	}

	api = &APICommon{
		mempool,
		chain,
		&atomic.Value{}, //*APIBlockchain
		&atomic.Value{}, //*APIBlockchainSync
		apiCommonFaucet,
		apiDelegatesNode,
		apiStore,
		&sync.Map{},
		&atomic.Value{},
	}

	api.MempoolProcessedThisBlock.Store(&sync.Map{})

	recovery.SafeGo(func() {

		updateNewChainDataUpdateListener := api.chain.UpdateNewChainDataUpdate.AddListener()
		defer api.chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainDataUpdateListener)

		for {
			newChainDataUpdateReceived, ok := <-updateNewChainDataUpdateListener
			if !ok {
				break
			}

			newChainDataUpdate := newChainDataUpdateReceived.(*blockchain.BlockchainDataUpdate)
			//it is safe to read
			api.readLocalBlockchain(newChainDataUpdate)

		}
	})

	recovery.SafeGo(func() {
		updateNewSync := api.chain.Sync.UpdateSyncMulticast.AddListener()
		defer api.chain.Sync.UpdateSyncMulticast.RemoveChannel(updateNewSync)

		for {
			newSyncDataReceived, ok := <-updateNewSync
			if !ok {
				break
			}

			newSyncData := newSyncDataReceived.(*blockchain_sync.BlockchainSyncData)
			api.readLocalBlockchainSync(newSyncData)
		}
	})

	api.readLocalBlockchain(chain.GetChainDataUpdate())

	return
}
