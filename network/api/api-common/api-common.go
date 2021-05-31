package api_common

import (
	"encoding/hex"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"sync/atomic"
)

type APICommon struct {
	mempool        *mempool.Mempool       `json:"-"`
	chain          *blockchain.Blockchain `json:"-"`
	localChain     *atomic.Value          `json:"-"` //*APIBlockchain
	localChainSync *atomic.Value          `json:"-"` //*APIBlockchain
	ApiStore       *APIStore              `json:"-"`
}

func (api *APICommon) GetBlockchain() (interface{}, error) {
	return api.localChain.Load().(*APIBlockchain), nil
}

func (api *APICommon) GetBlockchainSync() (interface{}, error) {
	return api.localChainSync.Load().(*APIBlockchainSync), nil
}

func (api *APICommon) GetInfo() (interface{}, error) {
	return &struct {
		Name       string `json:"name"`
		Version    string `json:"version"`
		Network    uint64 `json:"network"`
		CPUThreads int    `json:"CPUThreads"`
	}{
		Name:       config.NAME,
		Version:    config.VERSION,
		Network:    config.NETWORK_SELECTED,
		CPUThreads: config.CPU_THREADS,
	}, nil
}

func (api *APICommon) GetPing() (interface{}, error) {
	return &struct {
		Ping string `json:"ping"`
	}{Ping: "pong"}, nil
}

func (api *APICommon) GetBlockHash(blockHeight uint64) (interface{}, error) {
	return api.ApiStore.LoadBlockHash(blockHeight)
}

func (api *APICommon) GetTxHash(blockHeight uint64) (interface{}, error) {
	return api.ApiStore.LoadTxHash(blockHeight)
}

func (api *APICommon) GetBlockComplete(request *APIBlockCompleteRequest) (interface{}, error) {

	var blockComplete *block_complete.BlockComplete
	var err error

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(request.Height)
	}

	if err != nil {
		return nil, err
	}

	if request.ReturnType == RETURN_SERIALIZED {
		return blockComplete.SerializeToBytesBloomed(), nil
	}
	return blockComplete, nil
}

func (api *APICommon) GetBlock(request *APIBlockRequest) (interface{}, error) {
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		return api.ApiStore.LoadBlockWithTXsFromHash(request.Hash)
	}
	return api.ApiStore.LoadBlockWithTXsFromHeight(request.Height)
}

func (api *APICommon) GetBlockInfo(request *APIBlockRequest) (interface{}, error) {
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		return api.ApiStore.LoadBlockInfoFromHash(request.Hash)
	}
	return api.ApiStore.LoadBlockInfoFromHeight(request.Height)
}

func (api *APICommon) GetTx(request *APITransactionRequest) (out interface{}, err error) {

	var tx *transaction.Transaction

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		tx = api.mempool.Exists(request.Hash)
		if tx == nil {
			tx, err = api.ApiStore.LoadTxFromHash(request.Hash)
		}
	} else {
		tx, err = api.ApiStore.LoadTxFromHeight(request.Height)
	}

	if err != nil {
		return
	}

	if request.ReturnType == RETURN_SERIALIZED {
		out = &APITransactionSerialized{
			Tx:      tx.SerializeToBytesBloomed(),
			Mempool: tx != nil,
		}
	} else if request.ReturnType == RETURN_JSON {
		out = &APITransaction{
			Tx:      tx,
			Mempool: tx != nil,
		}
	} else {
		err = errors.New("Invalid return type")
	}

	return
}

func (api *APICommon) GetAccount(request *APIAccountRequest) (out interface{}, err error) {

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
	return acc, nil
}

func (api *APICommon) GetToken(hash []byte) (interface{}, error) {
	return api.ApiStore.LoadTokenFromPublicKeyHash(hash)
}

func (api *APICommon) GetMempool() (interface{}, error) {
	transactions := api.mempool.GetTxsList()
	hashes := make([]helpers.HexBytes, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Tx.Bloom.Hash
	}
	return hashes, nil
}

func (api *APICommon) GetMempoolExists(txId []byte) (interface{}, error) {
	if len(txId) != 32 {
		return nil, errors.New("TxId must be 32 byte")
	}
	return api.mempool.Exists(txId), nil
}

func (api *APICommon) PostMempoolInsert(tx *transaction.Transaction) (interface{}, error) {
	if err := tx.BloomAll(); err != nil {
		return nil, err
	}
	return api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true)
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
