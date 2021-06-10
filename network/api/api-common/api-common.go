package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
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
	localChainSync *atomic.Value //*APIBlockchain
	ApiStore       *APIStore
}

func (api *APICommon) GetBlockchain() ([]byte, error) {
	chain := api.localChain.Load().(*api_types.APIBlockchain)
	return json.Marshal(chain)
}

func (api *APICommon) GetBlockchainSync() ([]byte, error) {
	sync := api.localChainSync.Load().(*api_types.APIBlockchainSync)
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

func (api *APICommon) GetBlockComplete(request *api_types.APIBlockCompleteRequest) (out []byte, err error) {

	var blockComplete *block_complete.BlockComplete

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHeight(request.Height)
	}
	if err != nil || blockComplete == nil {
		return
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return blockComplete.SerializeToBytesBloomed(), nil
	}
	return json.Marshal(blockComplete)
}

func (api *APICommon) GetBlock(request *api_types.APIBlockRequest) (out []byte, err error) {
	var block interface{}
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		block, err = api.ApiStore.openLoadBlockWithTXsFromHash(request.Hash)
	} else {
		block, err = api.ApiStore.openLoadBlockWithTXsFromHeight(request.Height)
	}
	if err != nil || block == nil {
		return
	}
	return json.Marshal(block)
}

func (api *APICommon) GetBlockInfo(request *api_types.APIBlockRequest) (out []byte, err error) {
	var blockInfo *info.BlockInfo
	blockInfo, err = api.ApiStore.openLoadBlockInfo(request.Height, request.Hash)
	if err != nil || blockInfo == nil {
		return
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetTx(request *api_types.APITransactionRequest) (out []byte, err error) {

	var tx *transaction.Transaction

	mempool := false
	var txInfo *info.TxInfo
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		tx = api.mempool.Txs.Exists(request.Hash)
		if tx != nil {
			mempool = true
		} else {
			tx, txInfo, err = api.ApiStore.openLoadTx(request.Hash, request.Height)
		}
	} else {
		tx, txInfo, err = api.ApiStore.openLoadTx(request.Hash, request.Height)
	}

	if err != nil || tx == nil {
		return
	}

	result := &api_types.APITransaction{nil, nil, mempool, txInfo}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		result.TxSerialized = tx.SerializeToBytesBloomed()
	} else if request.ReturnType == api_types.RETURN_JSON {
		result.Tx = tx
	} else {
		return nil, errors.New("Invalid return type")
	}

	return json.Marshal(result)
}

func (api *APICommon) GetAccount(request *api_types.APIAccountRequest) (out []byte, err error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return
	}

	acc, err := api.ApiStore.openLoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil || acc == nil {
		return
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return acc.SerializeToBytes(), nil
	}
	return json.Marshal(acc)
}

func (api *APICommon) GetAccountTxs(request *api_types.APIAccountTxsRequest) (out []byte, err error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return
	}

	answer, err := api.ApiStore.openLoadAccountTxsFromPublicKeyHash(publicKeyHash, request.Next)
	if err != nil || answer == nil {
		return
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetTxInfo(request *api_types.APITransactionInfoRequest) (out []byte, err error) {
	txInfo, err := api.ApiStore.openLoadTxInfo(request.Hash, request.Height)
	if err != nil || txInfo == nil {
		return
	}
	return json.Marshal(txInfo)
}

func (api *APICommon) GetTokenInfo(request *api_types.APITokenInfoRequest) (out []byte, err error) {
	var tokInfo *info.TokenInfo
	if request.Hash != nil && (len(request.Hash) == cryptography.PublicKeyHashHashSize || len(request.Hash) == 0) {
		tokInfo, err = api.ApiStore.openLoadTokenInfo(request.Hash)
	}
	if err != nil || tokInfo == nil {
		return
	}
	return json.Marshal(tokInfo)
}

func (api *APICommon) GetToken(request *api_types.APITokenRequest) (out []byte, err error) {
	token, err := api.ApiStore.openLoadTokenFromPublicKeyHash(request.Hash)
	if err != nil || token == nil {
		return
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return token.SerializeToBytes(), nil
	}
	return json.Marshal(token)
}

func (api *APICommon) GetMempool(request *api_types.APIMempoolRequest) (out []byte, err error) {
	transactions := api.mempool.Txs.GetTxsList()

	length := len(transactions) - request.Start
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
		result.Hashes[i] = transactions[request.Start+i].Tx.Bloom.Hash
	}
	return json.Marshal(result)
}

func (api *APICommon) GetMempoolExists(txId []byte) (out []byte, err error) {
	if len(txId) != 32 {
		return nil, errors.New("TxId must be 32 byte")
	}
	tx := api.mempool.Txs.Exists(txId)
	if tx == nil {
		return nil, errors.New("Tx is not in mempool")
	}
	return json.Marshal(tx)
}

func (api *APICommon) PostMempoolInsert(tx *transaction.Transaction) (out []byte, err error) {
	if err = tx.BloomAll(); err != nil {
		return
	}
	result, err := api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true)
	if err != nil {
		return
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
func (api *APICommon) readLocalBlockchainSync(SyncTime uint64) {
	newLocalSync := &api_types.APIBlockchainSync{
		SyncTime: SyncTime,
	}
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
		for {
			newSyncDataReceived, ok := <-updateNewSync
			if !ok {
				break
			}

			newSyncData := newSyncDataReceived.(uint64)
			api.readLocalBlockchainSync(newSyncData)
		}
	})

	api.readLocalBlockchain(chain.GetChainDataUpdate())

	return
}
