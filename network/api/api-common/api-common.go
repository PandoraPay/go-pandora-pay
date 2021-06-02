package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	block_info "pandora-pay/blockchain/block-info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
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
	chain := api.localChain.Load().(*APIBlockchain)
	return json.Marshal(chain)
}

func (api *APICommon) GetBlockchainSync() ([]byte, error) {
	sync := api.localChainSync.Load().(*APIBlockchainSync)
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

func (api *APICommon) GetBlockComplete(request *APIBlockCompleteRequest) (out []byte, err error) {

	var blockComplete *block_complete.BlockComplete

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(request.Height)
	}
	if err != nil {
		return
	}

	if request.ReturnType == RETURN_SERIALIZED {
		return blockComplete.SerializeToBytesBloomed(), nil
	}
	return json.Marshal(blockComplete)
}

func (api *APICommon) GetBlock(request *APIBlockRequest) (out []byte, err error) {
	var block interface{}
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		block, err = api.ApiStore.LoadBlockWithTXsFromHash(request.Hash)
	} else {
		block, err = api.ApiStore.LoadBlockWithTXsFromHeight(request.Height)
	}
	if err != nil {
		return
	}
	return json.Marshal(block)
}

func (api *APICommon) GetBlockInfo(request *APIBlockRequest) (out []byte, err error) {
	var blockInfo *block_info.BlockInfo
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockInfo, err = api.ApiStore.LoadBlockInfoFromHash(request.Hash)
	} else {
		blockInfo, err = api.ApiStore.LoadBlockInfoFromHeight(request.Height)
	}
	if err != nil {
		return
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetTx(request *APITransactionRequest) (out []byte, err error) {

	var tx *transaction.Transaction

	mempool := false
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		tx = api.mempool.Exists(request.Hash)
		if tx != nil {
			mempool = true
		} else {
			tx, err = api.ApiStore.LoadTxFromHash(request.Hash)
		}
	} else {
		tx, err = api.ApiStore.LoadTxFromHeight(request.Height)
	}
	if err != nil {
		return
	}

	if request.ReturnType == RETURN_SERIALIZED {
		return json.Marshal(&APITransactionSerialized{
			Tx:      tx.SerializeToBytesBloomed(),
			Mempool: mempool,
		})
	} else if request.ReturnType == RETURN_JSON {
		return json.Marshal(&APITransaction{
			Tx:      tx,
			Mempool: mempool,
		})
	} else {
		return nil, errors.New("Invalid return type")
	}
}

func (api *APICommon) GetAccount(request *APIAccountRequest) (out []byte, err error) {

	var publicKeyHash []byte
	if request.Address != "" {
		address, err := addresses.DecodeAddr(request.Address)
		if err != nil {
			return nil, errors.New("Invalid address")
		}
		publicKeyHash = address.PublicKeyHash
	} else if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		publicKeyHash = request.Hash
	} else {
		return nil, errors.New("Invalid address")
	}

	acc, err := api.ApiStore.LoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil {
		return
	}

	if request.ReturnType == RETURN_SERIALIZED {
		return acc.SerializeToBytes(), nil
	}
	return json.Marshal(acc)
}

func (api *APICommon) GetToken(request *APITokenRequest) (out []byte, err error) {
	token, err := api.ApiStore.LoadTokenFromPublicKeyHash(request.Hash)
	if err != nil {
		return
	}
	if request.ReturnType == RETURN_SERIALIZED {
		return token.SerializeToBytes(), nil
	}
	return json.Marshal(token)
}

func (api *APICommon) GetMempool() (out []byte, err error) {
	transactions := api.mempool.GetTxsList()
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
	tx := api.mempool.Exists(txId)
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
	newLocalChain := &APIBlockchain{
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
	newLocalSync := &APIBlockchainSync{
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

		updateNewChainDataUpdateListener := api.chain.UpdateNewChainDataUpdateMulticast.AddListener()
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
