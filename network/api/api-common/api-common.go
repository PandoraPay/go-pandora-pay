package api_common

import (
	"encoding/hex"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"sync/atomic"
)

type APICommon struct {
	mempool    *mempool.Mempool
	chain      *blockchain.Blockchain
	localChain atomic.Value //*APIBlockchain
	ApiStore   *APIStore
}

func (api *APICommon) GetBlockchain() (interface{}, error) {
	return api.localChain.Load().(*APIBlockchain), nil
}

func (api *APICommon) GetInfo() (interface{}, error) {
	return &struct {
		Name        string
		Version     string
		Network     uint64
		CPU_THREADS int
	}{
		Name:        config.NAME,
		Version:     config.VERSION,
		Network:     config.NETWORK_SELECTED,
		CPU_THREADS: config.CPU_THREADS,
	}, nil
}

func (api *APICommon) GetPing() (interface{}, error) {
	return &struct {
		Ping string
	}{Ping: "Pong"}, nil
}

func (api *APICommon) GetBlockHash(blockHeight uint64) (interface{}, error) {
	return api.ApiStore.LoadBlockHash(blockHeight)
}

func (api *APICommon) GetBlockComplete(height uint64, hash []byte, typeValue uint8) (interface{}, error) {

	var blockComplete *block_complete.BlockComplete
	var err error

	if hash != nil {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHash(hash)
	} else {
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(height)
	}

	if err != nil {
		return nil, err
	}

	if typeValue == 1 {
		return blockComplete.Serialize(), nil
	}
	return blockComplete, nil
}

func (api *APICommon) GetBlock(height uint64, hash []byte) (interface{}, error) {
	if hash == nil {
		return api.ApiStore.LoadBlockWithTXsFromHash(hash)
	} else {
		return api.ApiStore.LoadBlockWithTXsFromHeight(height)
	}
}

func (api *APICommon) GetTx(hash []byte, typeValue uint8) (interface{}, error) {

	var err error

	var tx *transaction.Transaction
	output := struct {
		tx      interface{}
		mempool bool
	}{
		tx:      nil,
		mempool: false,
	}

	tx = api.mempool.Exists(hash)
	if tx != nil {
		output.mempool = true
	} else {
		tx, err = api.ApiStore.LoadTxFromHash(hash)
	}

	if err != nil {
		return nil, err
	}

	if typeValue == 1 {
		output.tx = tx.Serialize()
	} else {
		output.tx = tx
	}

	return output, nil
}

func (api *APICommon) GetBalance(address *addresses.Address, hash []byte) (interface{}, error) {
	if address != nil {
		return api.ApiStore.LoadAccountFromPublicKeyHash(address.PublicKeyHash)
	} else if hash != nil {
		return api.ApiStore.LoadAccountFromPublicKeyHash(hash)
	}
	return nil, errors.New("Invalid address or hash")
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
	return api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true)
}

//make sure it is safe to read
func (api *APICommon) readLocalBlockchain(newChainData *blockchain.BlockchainData) {
	newLocalChain := &APIBlockchain{
		Height:          newChainData.Height,
		Hash:            hex.EncodeToString(newChainData.Hash),
		PrevHash:        hex.EncodeToString(newChainData.PrevHash),
		KernelHash:      hex.EncodeToString(newChainData.KernelHash),
		PrevKernelHash:  hex.EncodeToString(newChainData.PrevKernelHash),
		Timestamp:       newChainData.Timestamp,
		Transactions:    newChainData.Transactions,
		Target:          newChainData.Target.String(),
		TotalDifficulty: newChainData.BigTotalDifficulty.String(),
	}
	api.localChain.Store(newLocalChain)
}

func CreateAPICommon(mempool *mempool.Mempool, chain *blockchain.Blockchain, apiStore *APIStore) (api *APICommon) {

	api = &APICommon{
		mempool,
		chain,
		atomic.Value{}, //*APIBlockchain
		apiStore,
	}

	go func() {
		updateNewChain := api.chain.UpdateNewChainMulticast.AddListener()
		for {
			newChainDataReceived, ok := <-updateNewChain
			if !ok {
				break
			}

			newChainData := newChainDataReceived.(*blockchain.BlockchainData)
			//it is safe to read
			api.readLocalBlockchain(newChainData)

		}
	}()

	api.readLocalBlockchain(chain.GetChainData())

	return
}
