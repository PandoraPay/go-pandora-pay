package api_common

import (
	"encoding/base64"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/config"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers/generics"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common/api_delegator_node"
	"pandora-pay/network/api/api_common/api_faucet"
	"pandora-pay/network/known_nodes"
	"pandora-pay/recovery"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"time"
)

type mempoolNewTxReply struct {
	wait   chan struct{}
	result bool
	err    error
}

type APICommon struct {
	mempool                   *mempool.Mempool
	txsValidator              *txs_validator.TxsValidator
	txsBuilder                *txs_builder.TxsBuilder
	chain                     *blockchain.Blockchain
	wallet                    *wallet.Wallet
	knownNodes                *known_nodes.KnownNodes
	localChain                *generics.Value[*APIBlockchain]
	localChainSync            *generics.Value[*blockchain_sync.BlockchainSyncData]
	Faucet                    *api_faucet.Faucet
	DelegatorNode             *api_delegator_node.DelegatorNode
	ApiStore                  *APIStore
	mempoolProcessedThisBlock *generics.Value[*generics.Map[string, *mempoolNewTxReply]]
	temporaryList             *generics.Value[*APINetworkNodesReply]
	temporaryListCreation     *generics.Value[time.Time]
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchain(newChainDataUpdate *blockchain.BlockchainDataUpdate) {
	newLocalChain := &APIBlockchain{
		newChainDataUpdate.Update.Height,
		base64.StdEncoding.EncodeToString(newChainDataUpdate.Update.Hash),
		base64.StdEncoding.EncodeToString(newChainDataUpdate.Update.PrevHash),
		base64.StdEncoding.EncodeToString(newChainDataUpdate.Update.KernelHash),
		base64.StdEncoding.EncodeToString(newChainDataUpdate.Update.PrevKernelHash),
		newChainDataUpdate.Update.Timestamp,
		newChainDataUpdate.Update.TransactionsCount,
		newChainDataUpdate.Update.AccountsCount,
		newChainDataUpdate.Update.AssetsCount,
		newChainDataUpdate.Update.Target.String(),
		newChainDataUpdate.Update.Supply,
		newChainDataUpdate.Update.BigTotalDifficulty.String(),
	}
	api.localChain.Store(newLocalChain)
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchainSync(newLocalSync *blockchain_sync.BlockchainSyncData) {
	api.localChainSync.Store(newLocalSync)
}

func NewAPICommon(knownNodes *known_nodes.KnownNodes, mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, txsValidator *txs_validator.TxsValidator, txsBuilder *txs_builder.TxsBuilder, apiStore *APIStore) (api *APICommon, err error) {

	var faucet *api_faucet.Faucet
	if config.NETWORK_SELECTED == config.TEST_NET_NETWORK_BYTE || config.NETWORK_SELECTED == config.DEV_NET_NETWORK_BYTE {
		if faucet, err = api_faucet.NewFaucet(mempool, chain, wallet, txsBuilder); err != nil {
			return
		}
	}

	var delegatorNode *api_delegator_node.DelegatorNode
	if config_nodes.DELEGATOR_ENABLED {
		delegatorNode = api_delegator_node.NewDelegatorNode(chain, wallet)
	}

	api = &APICommon{
		mempool,
		txsValidator,
		txsBuilder,
		chain,
		wallet,
		knownNodes,
		&generics.Value[*APIBlockchain]{},
		&generics.Value[*blockchain_sync.BlockchainSyncData]{},
		faucet,
		delegatorNode,
		apiStore,
		&generics.Value[*generics.Map[string, *mempoolNewTxReply]]{},
		&generics.Value[*APINetworkNodesReply]{},
		&generics.Value[time.Time]{},
	}

	api.temporaryListCreation.Store(time.Now())

	api.mempoolProcessedThisBlock.Store(&generics.Map[string, *mempoolNewTxReply]{})

	recovery.SafeGo(func() {

		updateNewChainDataUpdateListener := api.chain.UpdateNewChainDataUpdate.AddListener()
		defer api.chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainDataUpdateListener)

		for {
			newChainDataUpdate, ok := <-updateNewChainDataUpdateListener
			if !ok {
				return
			}

			//it is safe to read
			api.readLocalBlockchain(newChainDataUpdate)

			api.mempoolProcessedThisBlock.Store(&generics.Map[string, *mempoolNewTxReply]{})

		}
	})

	recovery.SafeGo(func() {
		updateNewSync := api.chain.Sync.UpdateSyncMulticast.AddListener()
		defer api.chain.Sync.UpdateSyncMulticast.RemoveChannel(updateNewSync)

		for {
			newSyncData, ok := <-updateNewSync
			if !ok {
				return
			}

			api.readLocalBlockchainSync(newSyncData)
		}
	})

	api.readLocalBlockchain(chain.GetChainDataUpdate())

	return
}
