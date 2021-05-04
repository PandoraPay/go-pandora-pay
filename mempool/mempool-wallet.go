package mempool

import (
	"sync"
)

type mempoolWalletAddress struct {
	publicKeyHash []byte
}

type mempoolWallet struct {
	myAddressesMap map[string]*mempoolWalletAddress
	sync.RWMutex   `json:"-"`
}

func (w *mempoolWallet) AddWallet(publicKeyHash []byte) {

	w.Lock()
	defer w.Unlock()

	w.myAddressesMap[string(publicKeyHash)] = &mempoolWalletAddress{
		publicKeyHash: publicKeyHash,
	}

}

func (w *mempoolWallet) RemoveWallet(publicKeyHash []byte) {

	w.Lock()
	defer w.Unlock()

	delete(w.myAddressesMap, string(publicKeyHash))
}

func createMempoolWallet() (w *mempoolWallet) {
	w = &mempoolWallet{
		myAddressesMap: make(map[string]*mempoolWalletAddress),
	}
	return
}
