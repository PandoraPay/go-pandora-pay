package api_http

import (
	"encoding/hex"
	"errors"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	api_store "pandora-pay/network/api/api-store"
	"pandora-pay/settings"
	"strconv"
	"sync/atomic"
	"unsafe"
)

type API struct {
	GetMap     map[string]func(values *url.Values) (interface{}, error)
	chain      *blockchain.Blockchain
	mempool    *mempool.Mempool
	localChain unsafe.Pointer
	ApiStore   *api_store.APIStore
}

func (api *API) getBlockchain(values *url.Values) (interface{}, error) {
	pointer := atomic.LoadPointer(&api.localChain)
	return (*APIBlockchain)(pointer), nil
}

func (api *API) getInfo(values *url.Values) (interface{}, error) {
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

func (api *API) getPing(values *url.Values) (interface{}, error) {
	return &struct {
		Ping string
	}{Ping: "Pong"}, nil
}

func (api *API) getBlockComplete(values *url.Values) (interface{}, error) {
	var blockComplete *block_complete.BlockComplete
	var err error

	if values.Get("height") != "" {
		var height uint64
		height, err = strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHeight(uint64(height))
	} else if values.Get("hash") != "" {
		var hash []byte
		hash, err = hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, errors.New("parameter 'hash' was is not a valid hex number")
		}
		blockComplete, err = api.ApiStore.LoadBlockCompleteFromHash(hash)
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}

	if err != nil {
		return nil, err
	}

	if values.Get("bytes") == "1" {
		return blockComplete.Serialize(), nil
	}
	return blockComplete, nil
}

func (api *API) getBlock(values *url.Values) (interface{}, error) {

	if values.Get("height") != "" {
		height, err := strconv.Atoi(values.Get("height"))
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		return api.ApiStore.LoadBlockWithTXsFromHeight(uint64(height))
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, errors.New("parameter 'hash' was is not a valid hex number")
		}
		return api.ApiStore.LoadBlockWithTXsFromHash(hash)
	}
	return nil, errors.New("parameter 'hash' or 'height' are missing")
}

func (api *API) getTx(values *url.Values) (interface{}, error) {
	var err error
	var tx *transaction.Transaction

	if values.Get("hash") != "" {
		var hash []byte
		hash, err = hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, errors.New("parameter 'hash' was is not a valid hex number")
		}
		tx, err = api.ApiStore.LoadTxFromHash(hash)
	} else {
		err = errors.New("parameter 'hash' was not specified ")
	}
	if err != nil {
		return nil, err
	}
	if values.Get("bytes") == "1" {
		return tx.Serialize(), nil
	}
	return tx, nil
}

func (api *API) getBalance(values *url.Values) (interface{}, error) {
	if values.Get("address") != "" {
		address, err := addresses.DecodeAddr(values.Get("address"))
		if err != nil {
			return nil, err
		}
		return api.ApiStore.LoadAccountFromPublicKeyHash(address.PublicKeyHash)
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, err
		}
		return api.ApiStore.LoadAccountFromPublicKeyHash(hash)
	}
	return nil, errors.New("parameter 'address' or 'hash' was not specified")
}

func (api *API) getToken(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.ApiStore.LoadTokenFromPublicKeyHash(hash)
}

func (api *API) getMempool(values *url.Values) (interface{}, error) {
	transactions := api.mempool.GetTxsList()
	hashes := make([]helpers.ByteString, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Tx.Bloom.Hash
	}
	return hashes, nil
}

//make sure it is safe to read
func (api *API) readLocalBlockchain(newChainData *blockchain.BlockchainData) {
	newLocalChain := APIBlockchain{
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
	atomic.StorePointer(&api.localChain, unsafe.Pointer(&newLocalChain))
}

func CreateAPI(apiStore *api_store.APIStore, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *API {

	api := API{
		chain:    chain,
		mempool:  mempool,
		ApiStore: apiStore,
	}

	api.GetMap = map[string]func(values *url.Values) (interface{}, error){
		"":               api.getInfo,
		"chain":          api.getBlockchain,
		"ping":           api.getPing,
		"block-complete": api.getBlockComplete,
		"block":          api.getBlock,
		"tx":             api.getTx,
		"balance":        api.getBalance,
		"token":          api.getToken,
		"mempool":        api.getMempool,
	}

	go func() {
		for {
			newChainData, ok := <-api.chain.UpdateNewChainChannel
			if ok {
				//it is safe to read
				api.readLocalBlockchain(newChainData)
			}
		}
	}()

	api.readLocalBlockchain(chain.GetChainData())

	return &api
}
