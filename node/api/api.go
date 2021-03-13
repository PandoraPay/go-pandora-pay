package api

import (
	"encoding/hex"
	"net/http"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"strconv"
	"sync/atomic"
	"unsafe"
)

type API struct {
	GetMap map[string]func(req *http.Request) interface{}

	chain      *blockchain.Blockchain
	localChain unsafe.Pointer
}

func (api *API) getBlockchain(req *http.Request) interface{} {
	pointer := atomic.LoadPointer(&api.localChain)
	return (*APIBlockchain)(pointer)
}

func (api *API) getInfo(req *http.Request) interface{} {
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

func (api *API) getPing(req *http.Request) interface{} {
	return &struct {
		Ping string
	}{Ping: "Pong"}
}

func (api *API) getBlock(req *http.Request) interface{} {
	height, err := strconv.Atoi(req.URL.Query().Get("height"))
	if err != nil {
		panic("Parameter Height was not specified or is not a number")
	}
	return api.chain.LoadBlockCompleteFromHeight(uint64(height))
}

func (api *API) getBlockByHash(req *http.Request) interface{} {
	hash, err := hex.DecodeString(req.URL.Query().Get("hash"))
	if err != nil {
		panic("Parameter Height was not specified or is not a valid hex number")
	}
	return api.chain.LoadBlockCompleteFromHash(hash)
}

//make sure it is safe to read
func (api *API) readLocalBlockchain(newChain *blockchain.Blockchain) {
	newLocalChain := APIBlockchain{
		Height:          newChain.Height,
		Hash:            hex.EncodeToString(newChain.Hash),
		PrevHash:        hex.EncodeToString(newChain.PrevHash),
		KernelHash:      hex.EncodeToString(newChain.KernelHash),
		PrevKernelHash:  hex.EncodeToString(newChain.PrevKernelHash),
		Timestamp:       newChain.Timestamp,
		Transactions:    newChain.Transactions,
		Target:          newChain.Target.String(),
		TotalDifficulty: newChain.BigTotalDifficulty.String(),
	}
	atomic.StorePointer(&api.localChain, unsafe.Pointer(&newLocalChain))
}

func CreateAPI(chain *blockchain.Blockchain) *API {

	api := API{
		chain: chain,
	}

	api.GetMap = map[string]func(req *http.Request) interface{}{
		"/":              api.getInfo,
		"/chain":         api.getBlockchain,
		"/ping":          api.getPing,
		"/block":         api.getBlock,
		"/block-by-hash": api.getBlockByHash,
	}

	go func() {
		for {
			newChain := <-api.chain.UpdateNewChainChannel
			//it is safe to read
			api.readLocalBlockchain(newChain)
		}
	}()

	chain.RLock()
	api.readLocalBlockchain(chain)
	chain.RUnlock()

	return &api
}
