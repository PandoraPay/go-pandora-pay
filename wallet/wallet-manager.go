package wallet

import (
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

type Wallet struct {
	Version       Version
	Mnemonic      string
	Seed          [32]byte
	SequenceIndex uint32
	Count         int
	Addresses     []*WalletAddress `json:"-"`
}

var wallet Wallet

func addNewAddress() error {

	masterKey, _ := bip32.NewMasterKey(wallet.Seed[:])
	key, err := masterKey.NewChildKey(wallet.SequenceIndex)
	if err != nil {
		gui.Fatal("Couldn't derivate the marker key")
	}

	privateKey := addresses.PrivateKey{Key: key.Key}
	address := privateKey.GenerateTransparentAddress(true, 0, &[]byte{})

	walletAddress := WalletAddress{Version: WalletAddressTransparent, Name: "Addr " + strconv.Itoa(wallet.Count), PrivateKey: &privateKey, Address: address, IsSeedGenerated: true}

	wallet.Addresses = append(wallet.Addresses, &walletAddress)
	wallet.Count += 1
	wallet.SequenceIndex += 1

	updateWallet()
	return saveWallet()
}

func removeAddress(index int) error {
	if index < 0 || index > len(wallet.Addresses) {
		return errors.New("Invalid Address Index")
	}
	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	updateWallet()
	return saveWallet()
}

func showPrivateKey(index int) ([]byte, error) {
	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key, nil
}

func createSeed() error {

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return gui.Error("Entropy of the address raised an error", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return gui.Error("Mnemonic couldn't be created", err)
	}
	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = *helpers.Byte32(seed)

	return nil
}

func createEmptyWallet() error {
	wallet = Wallet{}

	err := createSeed()
	if err != nil {
		return gui.Error("Error creating seed", err)
	}
	return addNewAddress()
}

func updateWallet() {
	gui.InfoUpdate("Wallet Encrypted", walletSaved.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}
