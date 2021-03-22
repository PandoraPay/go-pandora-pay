package api_http

import (
	"encoding/hex"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
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
	GetMap     map[string]func(values *url.Values) interface{}
	chain      *blockchain.Blockchain
	mempool    *mempool.Mempool
	localChain unsafe.Pointer
	apiStore   *api_store.APIStore
}

func (api *API) getBlockchain(values *url.Values) interface{} {
	pointer := atomic.LoadPointer(&api.localChain)
	return (*APIBlockchain)(pointer)
}

func (api *API) getInfo(values *url.Values) interface{} {
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
	}
}

func (api *API) getPing(values *url.Values) interface{} {
	return &struct {
		Ping string
	}{Ping: "Pong"}
}

func (api *API) getBlockComplete(values *url.Values) interface{} {
	if values.Get("height") != "" {
		height, err := strconv.Atoi(values.Get("height"))
		if err != nil {
			panic("parameter 'height' is not a number")
		}
		return api.apiStore.LoadBlockCompleteFromHeight(uint64(height))
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			panic("parameter 'hash' was is not a valid hex number")
		}
		return api.apiStore.LoadBlockCompleteFromHash(hash)
	}
	panic("parameter 'hash' or 'height' are missing")
}

func (api *API) getBlock(values *url.Values) interface{} {
	if values.Get("height") != "" {
		height, err := strconv.Atoi(values.Get("height"))
		if err != nil {
			panic("parameter 'height' is not a number")
		}
		return api.apiStore.LoadBlockWithTXsFromHeight(uint64(height))
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			panic("parameter 'hash' was is not a valid hex number")
		}
		return api.apiStore.LoadBlockWithTXsFromHash(hash)
	}
	panic("parameter 'hash' or 'height' are missing")
}

func (api *API) getTx(values *url.Values) interface{} {
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			panic("parameter 'hash' was is not a valid hex number")
		}
		return api.apiStore.LoadTxFromHash(hash)
	}
	panic("parameter 'hash' was not specified ")
}

func (api *API) getBalance(values *url.Values) interface{} {
	if values.Get("address") != "" {
		address := addresses.DecodeAddr(values.Get("address"))
		return api.apiStore.LoadAccountFromPublicKeyHash(address.PublicKeyHash)
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			panic(err)
		}
		return api.apiStore.LoadAccountFromPublicKeyHash(hash)
	}
	panic("parameter 'address' or 'hash' was not specified")
}

func (api *API) getToken(values *url.Values) interface{} {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		panic(err)
	}
	return api.apiStore.LoadTokenFromPublicKeyHash(hash)
}

func (api *API) getMempool(values *url.Values) interface{} {
	transactions := api.mempool.GetTxsList()
	hashes := make([]helpers.ByteString, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Tx.Bloom.Hash
	}
	return hashes
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
		apiStore: apiStore,
	}

	api.GetMap = map[string]func(values *url.Values) interface{}{
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
