package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/blocks/block-info"
	token_info "pandora-pay/blockchain/tokens/token-info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common/api_types"
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
	return api.ApiStore.LoadBlockHash(blockHeight)
}

func (api *APICommon) GetTxHash(blockHeight uint64) (helpers.HexBytes, error) {
	return api.ApiStore.LoadTxHash(blockHeight)
}

func (api *APICommon) GetBlockComplete(request *api_types.APIBlockCompleteRequest) (out []byte, err error) {

	var blockComplete *block_complete.BlockComplete

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(request.Height)
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
		block, err = api.ApiStore.LoadBlockWithTXsFromHash(request.Hash)
	} else {
		block, err = api.ApiStore.LoadBlockWithTXsFromHeight(request.Height)
	}
	if err != nil || block == nil {
		return
	}
	return json.Marshal(block)
}

func (api *APICommon) GetBlockInfo(request *api_types.APIBlockRequest) (out []byte, err error) {
	var blockInfo *block_info.BlockInfo
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockInfo, err = api.ApiStore.LoadBlockInfoFromHash(request.Hash)
	} else {
		blockInfo, err = api.ApiStore.LoadBlockInfoFromHeight(request.Height)
	}
	if err != nil || blockInfo == nil {
		return
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetTx(request *api_types.APITransactionRequest) (out []byte, err error) {

	var tx *transaction.Transaction

	mempool := false
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		tx = api.mempool.Txs.Exists(request.Hash)
		if tx != nil {
			mempool = true
		} else {
			tx, err = api.ApiStore.LoadTxFromHash(request.Hash)
		}
	} else {
		tx, err = api.ApiStore.LoadTxFromHeight(request.Height)
	}
	if err != nil || tx == nil {
		return
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return json.Marshal(&api_types.APITransactionSerialized{
			Tx:      tx.SerializeToBytesBloomed(),
			Mempool: mempool,
		})
	} else if request.ReturnType == api_types.RETURN_JSON {
		return json.Marshal(&api_types.APITransaction{
			Tx:      tx,
			Mempool: mempool,
		})
	} else {
		return nil, errors.New("Invalid return type")
	}
}

func (api *APICommon) GetAccount(request *api_types.APIAccountRequest) (out []byte, err error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return
	}

	acc, err := api.ApiStore.LoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil || acc == nil {
		return
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return acc.SerializeToBytes(), nil
	}
	return json.Marshal(acc)
}

func (api *APICommon) GetTokenInfo(request *api_types.APITokenInfoRequest) (out []byte, err error) {
	var tokInfo *token_info.TokenInfo
	if request.Hash != nil && len(request.Hash) == cryptography.PublicKeyHashHashSize {
		tokInfo, err = api.ApiStore.LoadTokenInfoFromHash(request.Hash)
	}
	if err != nil || tokInfo == nil {
		return
	}
	return json.Marshal(tokInfo)
}

func (api *APICommon) GetToken(request *api_types.APITokenRequest) (out []byte, err error) {
	token, err := api.ApiStore.LoadTokenFromPublicKeyHash(request.Hash)
	if err != nil || token == nil {
		return
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return token.SerializeToBytes(), nil
	}
	return json.Marshal(token)
}

func (api *APICommon) GetMempool() (out []byte, err error) {
	transactions := api.mempool.Txs.GetTxsList()
	hashes := make([]helpers.HexBytes, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Tx.Bloom.Hash
	}
	return json.Marshal(hashes)
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

	go func() {

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
	}()

	go func() {
		updateNewSync := api.chain.Sync.UpdateSyncMulticast.AddListener()
		for {
			newSyncDataReceived, ok := <-updateNewSync
			if !ok {
				break
			}

			newSyncData := newSyncDataReceived.(uint64)
			api.readLocalBlockchainSync(newSyncData)
		}
	}()

	api.readLocalBlockchain(chain.GetChainDataUpdate())

	return
}
