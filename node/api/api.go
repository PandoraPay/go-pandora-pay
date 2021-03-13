package api

import (
	"encoding/hex"
	"net/http"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"sync/atomic"
	"unsafe"
)

type API struct {
	GetMap map[string]func(req *http.Request) interface{}

	chain      *blockchain.Blockchain
	localChain unsafe.Pointer
}

func (api *API) blockchain(req *http.Request) interface{} {
	pointer := atomic.LoadPointer(&api.localChain)
	localchain := (*APIBlockchain)(pointer)

	return &localchain
}

func (api *API) info(req *http.Request) interface{} {
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

func (api *API) ping(req *http.Request) interface{} {
	return &struct {
		Ping string
	}{Ping: "Pong"}
}

//make sure it is safe to read
func (api *API) loadLocalBlockchain(newChain *blockchain.Blockchain) {
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
		"/":           api.info,
		"/blockchain": api.blockchain,
		"/ping":       api.ping,
	}

	go func() {
		for {
			newChain := <-api.chain.UpdateNewChainChannel
			//it is safe to read
			api.loadLocalBlockchain(newChain)
		}
	}()

	chain.RLock()
	api.loadLocalBlockchain(chain)
	chain.RUnlock()

	return &api
}
