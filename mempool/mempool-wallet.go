package mempool

import (
	"pandora-pay/helpers"
	"sync"
)

type mempoolWalletAddress struct {
	publicKeyHash helpers.HexBytes `json:"-"`
}

type mempoolWallet struct {
	myAddressesMap map[string]*mempoolWalletAddress `json:"-"`
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

func (w *mempoolWallet) Exists(publicKeyHash []byte) bool {

	w.RLock()
	defer w.RUnlock()

	return w.myAddressesMap[string(publicKeyHash)] != nil
}

func createMempoolWallet() (w *mempoolWallet) {
	w = &mempoolWallet{
		myAddressesMap: make(map[string]*mempoolWalletAddress),
	}
	return
}
