package api_common

import (
	"encoding/hex"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/config"
	"pandora-pay/config/config_nodes"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common/api_delegator_node"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/known_nodes"
	"pandora-pay/recovery"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type APICommon struct {
	mempool                   *mempool.Mempool
	chain                     *blockchain.Blockchain
	knownNodes                *known_nodes.KnownNodes
	localChain                *atomic.Value //*APIBlockchain
	localChainSync            *atomic.Value //*blockchain_sync.BlockchainSyncData
	Faucet                    *api_faucet.Faucet
	DelegatorNode             *api_delegator_node.DelegatorNode
	ApiStore                  *APIStore
	MempoolDownloadPending    *sync.Map     //[string]*mempoolNewTxAnswer
	MempoolProcessedThisBlock *atomic.Value // *sync.Map //[string]*APIMempoolNewTxReply

	temporaryList         atomic.Value //[]*KnownNode
	temporaryListCreation atomic.Value //time.Time
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

func NewAPICommon(knownNodes *known_nodes.KnownNodes, mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder, apiStore *APIStore) (api *APICommon, err error) {

	var faucet *api_faucet.Faucet
	if config.NETWORK_SELECTED == config.TEST_NET_NETWORK_BYTE || config.NETWORK_SELECTED == config.DEV_NET_NETWORK_BYTE {
		if faucet, err = api_faucet.NewFaucet(mempool, chain, wallet, transactionsBuilder); err != nil {
			return
		}
	}

	var delegatorNode *api_delegator_node.DelegatorNode
	if config_nodes.DELEGATES_ALLOWED_ENABLED {
		delegatorNode = api_delegator_node.NewDelegatorNode(chain, wallet)
	}

	api = &APICommon{
		mempool,
		chain,
		knownNodes,
		&atomic.Value{}, //*APIBlockchain
		&atomic.Value{}, //*APIBlockchainSync
		faucet,
		delegatorNode,
		apiStore,
		&sync.Map{},
		&atomic.Value{},
		atomic.Value{},
		atomic.Value{},
	}

	api.temporaryListCreation.Store(time.Now())

	api.MempoolProcessedThisBlock.Store(&sync.Map{})

	recovery.SafeGo(func() {

		updateNewChainDataUpdateListener := api.chain.UpdateNewChainDataUpdate.AddListener()
		defer api.chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainDataUpdateListener)

		for {
			newChainDataUpdateReceived, ok := <-updateNewChainDataUpdateListener
			if !ok {
				return
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
				return
			}

			newSyncData := newSyncDataReceived.(*blockchain_sync.BlockchainSyncData)
			api.readLocalBlockchainSync(newSyncData)
		}
	})

	api.readLocalBlockchain(chain.GetChainDataUpdate())

	return
}
