package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	blockchain_sync "pandora-pay/blockchain/blockchain-sync"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/recovery"
	"sync/atomic"
)

type APICommon struct {
	mempool        *mempool.Mempool
	chain          *blockchain.Blockchain
	localChain     *atomic.Value //*APIBlockchain
	localChainSync *atomic.Value //*blockchain_sync.BlockchainSyncData
	ApiStore       *APIStore
}

func (api *APICommon) GetBlockchain() ([]byte, error) {
	chain := api.localChain.Load().(*api_types.APIBlockchain)
	return json.Marshal(chain)
}

func (api *APICommon) GetBlockchainSync() ([]byte, error) {
	sync := api.localChainSync.Load().(*blockchain_sync.BlockchainSyncData)
	return json.Marshal(sync)
}

func (api *APICommon) GetInfo() ([]byte, error) {
	return json.Marshal(&struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Network    uint64 `json:"network"`
		CPUThreads int    `json:"CPUThreads"`
	}{
		Name:       config.NAME,
		Version:    config.VERSION,
		Network:    config.NETWORK_SELECTED,
		CPUThreads: config.CPU_THREADS,
	})
}

func (api *APICommon) GetPing() ([]byte, error) {
	return json.Marshal(&struct {
		Ping string `json:"ping"`
	}{Ping: "pong"})
}

func (api *APICommon) GetBlockHash(blockHeight uint64) (helpers.HexBytes, error) {
	return api.ApiStore.chain.OpenLoadBlockHash(blockHeight)
}

func (api *APICommon) GetTxHash(blockHeight uint64) (helpers.HexBytes, error) {
	return api.ApiStore.openLoadTxHash(blockHeight)
}

func (api *APICommon) GetBlockCompleteMissingTxs(request *api_types.APIBlockCompleteMissingTxsRequest) ([]byte, error) {

	var blockCompleteMissingTxs *api_types.APIBlockCompleteMissingTxs
	var err error

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockCompleteMissingTxs, err = api.ApiStore.openLoadBlockCompleteMissingTxsFromHash(request.Hash, request.MissingTxs)
	}
	if err != nil || blockCompleteMissingTxs == nil {
		return nil, err
	}
	return json.Marshal(blockCompleteMissingTxs)
}

func (api *APICommon) GetBlockComplete(request *api_types.APIBlockCompleteRequest) ([]byte, error) {

	var blockComplete *block_complete.BlockComplete
	var err error

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHeight(request.Height)
	}
	if err != nil || blockComplete == nil {
		return nil, err
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return blockComplete.BloomBlkComplete.Serialized, nil
	}
	return json.Marshal(blockComplete)
}

func (api *APICommon) GetBlock(request *api_types.APIBlockRequest) ([]byte, error) {

	var out *api_types.APIBlockWithTxs

	var err error
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		out, err = api.ApiStore.openLoadBlockWithTXsFromHash(request.Hash)
	} else {
		out, err = api.ApiStore.openLoadBlockWithTXsFromHeight(request.Height)
	}
	if err != nil || out.Block == nil {
		return nil, err
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		out.BlockSerialized = out.Block.SerializeToBytes()
		out.Block = nil
	}

	return json.Marshal(out)
}

func (api *APICommon) GetBlockInfo(request *api_types.APIBlockInfoRequest) ([]byte, error) {
	blockInfo, err := api.ApiStore.openLoadBlockInfo(request.Height, request.Hash)
	if err != nil || blockInfo == nil {
		return nil, err
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetTx(request *api_types.APITransactionRequest) ([]byte, error) {

	var tx *transaction.Transaction
	var err error

	mempool := false
	var txInfo *info.TxInfo
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		tx = api.mempool.Txs.Exists(string(request.Hash))
		if tx != nil {
			mempool = true
		} else {
			tx, txInfo, err = api.ApiStore.openLoadTx(request.Hash, request.Height)
		}
	} else {
		tx, txInfo, err = api.ApiStore.openLoadTx(request.Hash, request.Height)
	}

	if err != nil || tx == nil {
		return nil, err
	}

	result := &api_types.APITransaction{nil, nil, mempool, txInfo}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		result.TxSerialized = tx.Bloom.Serialized
	} else if request.ReturnType == api_types.RETURN_JSON {
		result.Tx = tx
	} else {
		return nil, errors.New("Invalid return type")
	}

	return json.Marshal(result)
}

func (api *APICommon) GetAccount(request *api_types.APIAccountRequest) ([]byte, error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return nil, err
	}

	acc, err := api.ApiStore.openLoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil || acc == nil {
		return nil, err
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return acc.SerializeToBytes(), nil
	}
	return json.Marshal(acc)
}

func (api *APICommon) GetAccountTxs(request *api_types.APIAccountTxsRequest) ([]byte, error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return nil, err
	}

	answer, err := api.ApiStore.openLoadAccountTxsFromPublicKeyHash(publicKeyHash, request.Next)
	if err != nil || answer == nil {
		return nil, err
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetTxInfo(request *api_types.APITransactionInfoRequest) ([]byte, error) {
	txInfo, err := api.ApiStore.openLoadTxInfo(request.Hash, request.Height)
	if err != nil || txInfo == nil {
		return nil, err
	}
	return json.Marshal(txInfo)
}

func (api *APICommon) GetTokenInfo(request *api_types.APITokenInfoRequest) ([]byte, error) {
	var tokInfo *info.TokenInfo
	var err error
	if request.Hash != nil && (len(request.Hash) == cryptography.PublicKeyHashHashSize || len(request.Hash) == 0) {
		tokInfo, err = api.ApiStore.openLoadTokenInfo(request.Hash)
	}
	if err != nil || tokInfo == nil {
		return nil, err
	}
	return json.Marshal(tokInfo)
}

func (api *APICommon) GetToken(request *api_types.APITokenRequest) ([]byte, error) {
	token, err := api.ApiStore.openLoadTokenFromPublicKeyHash(request.Hash)
	if err != nil || token == nil {
		return nil, err
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return token.SerializeToBytes(), nil
	}
	return json.Marshal(token)
}

func (api *APICommon) GetMempool(request *api_types.APIMempoolRequest) ([]byte, error) {
	transactions := api.mempool.Txs.GetTxsList()

	if request.Count == 0 {
		request.Count = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	start := request.Page * request.Count

	length := len(transactions) - start
	if length < 0 {
		length = 0
	}
	if length > config.API_MEMPOOL_MAX_TRANSACTIONS {
		length = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	result := &api_types.APIMempoolAnswer{
		Count:  len(transactions),
		Hashes: make([]helpers.HexBytes, length),
	}

	for i := 0; i < len(result.Hashes); i++ {
		result.Hashes[i] = transactions[start+i].Tx.Bloom.Hash
	}
	return json.Marshal(result)
}

func (api *APICommon) GetMempoolExists(txId []byte) ([]byte, error) {
	if len(txId) != 32 {
		return nil, errors.New("TxId must be 32 byte")
	}
	tx := api.mempool.Txs.Exists(string(txId))
	if tx == nil {
		return nil, errors.New("Tx is not in mempool")
	}
	return json.Marshal(tx)
}

func (api *APICommon) PostMempoolInsert(tx *transaction.Transaction) ([]byte, error) {
	if err := tx.BloomAll(); err != nil {
		return nil, err
	}
	result, err := api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true, true)
	if err != nil {
		return nil, err
	}
	if result {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchain(newChainDataUpdate *blockchain.BlockchainDataUpdate) {
	newLocalChain := &api_types.APIBlockchain{
		Height:            newChainDataUpdate.Update.Height,
		Hash:              hex.EncodeToString(newChainDataUpdate.Update.Hash),
		PrevHash:          hex.EncodeToString(newChainDataUpdate.Update.PrevHash),
		KernelHash:        hex.EncodeToString(newChainDataUpdate.Update.KernelHash),
		PrevKernelHash:    hex.EncodeToString(newChainDataUpdate.Update.PrevKernelHash),
		Timestamp:         newChainDataUpdate.Update.Timestamp,
		TransactionsCount: newChainDataUpdate.Update.TransactionsCount,
		Target:            newChainDataUpdate.Update.Target.String(),
		TotalDifficulty:   newChainDataUpdate.Update.BigTotalDifficulty.String(),
	}
	api.localChain.Store(newLocalChain)
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchainSync(newLocalSync *blockchain_sync.BlockchainSyncData) {
	api.localChainSync.Store(newLocalSync)
}

func CreateAPICommon(mempool *mempool.Mempool, chain *blockchain.Blockchain, apiStore *APIStore) (api *APICommon) {

	api = &APICommon{
		mempool,
		chain,
		&atomic.Value{}, //*APIBlockchain
		&atomic.Value{}, //*APIBlockchainSync
		apiStore,
	}

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
