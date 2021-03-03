package wallet

import (
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/forging"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
	"sync"
)

type Wallet struct {
	Version   Version
	Mnemonic  string
	Seed      [32]byte
	SeedIndex uint32
	Count     int
	Addresses []*WalletAddress `json:"-"`

	// forging creates multiple threads and it will read the wallet.Addresses
	sync.RWMutex `json:"-"`
}

var W Wallet

func (W *Wallet) addNewAddress() (err error) {

	masterKey, _ := bip32.NewMasterKey(W.Seed[:])

	var key *bip32.Key
	if key, err = masterKey.NewChildKey(W.SeedIndex); err != nil {
		gui.Fatal("Couldn't derivate the marker key", err)
	}

	privateKey := addresses.PrivateKey{Key: *helpers.Byte32(key.Key)}

	var publicKey [33]byte
	if publicKey, err = privateKey.GeneratePublicKey(); err != nil {
		gui.Fatal("Generating Public Key from Private key raised an error", err)
	}

	var address *addresses.Address
	if address, err = privateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		gui.Fatal("Generating Address raised an error", err)
	}

	publicKeyHash := crypto.ComputePublicKeyHash(publicKey)

	W.Lock()
	defer W.Unlock()
	walletAddress := WalletAddress{
		"Addr " + strconv.Itoa(W.Count),
		&privateKey,
		publicKey,
		publicKeyHash,
		address,
		W.SeedIndex,
	}

	W.Addresses = append(W.Addresses, &walletAddress)
	W.Count += 1
	W.SeedIndex += 1

	go forging.ForgingW.AddWallet(publicKey, privateKey.Key, publicKeyHash)

	updateWallet()
	return saveWallet()
}

func (W *Wallet) removeAddress(index int) error {

	W.Lock()
	defer W.Unlock()

	if index < 0 || index > len(W.Addresses) {
		return errors.New("Invalid Address Index")
	}

	removing := W.Addresses[index]

	W.Addresses = append(W.Addresses[:index], W.Addresses[index+1:]...)
	W.Count -= 1

	go forging.ForgingW.RemoveWallet(removing.PublicKey)

	updateWallet()
	return saveWallet()
}

func (W *Wallet) showPrivateKey(index int) (*[32]byte, error) {

	W.RLock()
	defer W.RUnlock()

	if index < 0 || index > len(W.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return &W.Addresses[index].PrivateKey.Key, nil
}

func (W *Wallet) createSeed() (err error) {

	W.Lock()
	defer W.Unlock()

	var entropy []byte
	if entropy, err = bip39.NewEntropy(256); err != nil {
		return gui.Error("Entropy of the address raised an error", err)
	}

	var mnemonic string
	if mnemonic, err = bip39.NewMnemonic(entropy); err != nil {
		return gui.Error("Mnemonic couldn't be created", err)
	}

	W.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	W.Seed = *helpers.Byte32(seed)

	return nil
}

func (W *Wallet) createEmptyWallet() error {
	if err := W.createSeed(); err != nil {
		return gui.Error("Error creating seed", err)
	}
	return W.addNewAddress()
}

func updateWallet() {
	gui.InfoUpdate("Wallet", wSaved.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(W.Count))
}
