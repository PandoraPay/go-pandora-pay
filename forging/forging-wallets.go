package forging

import (
	"bytes"
	"pandora-pay/addresses"
	"sync"
)

type ForgingWallets struct {
	addresses []*ForgingWalletAddress
	sync.RWMutex
}

type ForgingWalletAddress struct {
	privateKey *addresses.PrivateKey
	publicKey  [33]byte
}

var ForgingW = ForgingWallets{}

func (w *ForgingWallets) AddWallet(pub [33]byte, priv [32]byte) {

	w.Lock()
	defer w.Unlock()

	//make a clone to be memory safe
	var privateKey [32]byte
	var publicKey [33]byte

	copy(privateKey[:], priv[:])
	copy(publicKey[:], pub[:])

	private := addresses.PrivateKey{Key: privateKey}

	address := ForgingWalletAddress{publicKey: publicKey, privateKey: &private}
	w.addresses = append(w.addresses, &address)

}

func (w *ForgingWallets) RemoveWallet(publicKey [33]byte) {

	w.Lock()
	defer w.Unlock()

	for i, address := range w.addresses {
		if bytes.Equal(address.publicKey[:], publicKey[:]) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			return
		}
	}

}
