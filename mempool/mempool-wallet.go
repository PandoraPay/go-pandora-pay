package mempool

import (
	"pandora-pay/helpers"
	"sync"
)

type mempoolWalletAddress struct {
	publicKey helpers.HexBytes `json:"-"`
}

type mempoolWallet struct {
	myAddressesMap map[string]*mempoolWalletAddress `json:"-"`
	sync.RWMutex   `json:"-"`
}

func (w *mempoolWallet) AddWallet(publicKey []byte) {

	w.Lock()
	defer w.Unlock()

	w.myAddressesMap[string(publicKey)] = &mempoolWalletAddress{
		publicKey: publicKey,
	}

}

func (w *mempoolWallet) RemoveWallet(publicKey []byte) {

	w.Lock()
	defer w.Unlock()

	delete(w.myAddressesMap, string(publicKey))
}

func (w *mempoolWallet) Exists(publicKey []byte) bool {

	w.RLock()
	defer w.RUnlock()

	return w.myAddressesMap[string(publicKey)] != nil
}

func createMempoolWallet() (w *mempoolWallet) {
	w = &mempoolWallet{
		myAddressesMap: make(map[string]*mempoolWalletAddress),
	}
	return
}
