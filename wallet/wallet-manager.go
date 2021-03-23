package wallet

import (
	"bytes"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

func (wallet *Wallet) GetWalletAddressByAddress(addressEncoded string) (*WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	for _, addr := range wallet.Addresses {
		if bytes.Equal(addr.PublicKeyHash, address.PublicKeyHash) {
			return addr, nil
		}
	}

	return nil, errors.New("address was not found")
}

func (wallet *Wallet) AddNewAddress() (walletAddress *WalletAddress, err error) {

	//avoid generating the same address twice
	wallet.Lock()
	defer wallet.Unlock()

	masterKey, _ := bip32.NewMasterKey(wallet.Seed)

	key, err := masterKey.NewChildKey(wallet.SeedIndex)
	if err != nil {
		return
	}

	walletAddress = &WalletAddress{
		Name:       "Addr " + strconv.Itoa(wallet.Count),
		PrivateKey: &addresses.PrivateKey{Key: key.Key},
		SeedIndex:  wallet.SeedIndex,
	}
	if walletAddress.PublicKey, err = walletAddress.PrivateKey.GeneratePublicKey(); err != nil {
		return
	}
	if walletAddress.Address, err = walletAddress.PrivateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		return
	}
	walletAddress.AddressEncoded = walletAddress.Address.EncodeAddr()
	publicKeyHash := cryptography.ComputePublicKeyHash(walletAddress.PublicKey)

	wallet.Addresses = append(wallet.Addresses, walletAddress)
	wallet.Count += 1
	wallet.SeedIndex += 1

	go wallet.forging.Wallet.AddWallet(walletAddress.PublicKey, walletAddress.PrivateKey.Key, publicKeyHash)

	wallet.updateWallet()
	wallet.saveWallet(wallet.Count-1, wallet.Count, -1)

	return
}

func (wallet *Wallet) RemoveAddress(index int) (bool, error) {

	wallet.Lock()
	defer wallet.Unlock()

	if index < 0 || index > len(wallet.Addresses) {
		return false, errors.New("Invalid Address Index")
	}

	removing := wallet.Addresses[index]

	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	go wallet.forging.Wallet.RemoveWallet(removing.PublicKeyHash)

	wallet.updateWallet()
	wallet.saveWallet(index, wallet.Count, wallet.Count)
	return true, nil
}

func (wallet *Wallet) ShowPrivateKey(index int) ([]byte, error) { //32 byte

	wallet.RLock()
	defer wallet.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key, nil
}

func (wallet *Wallet) createSeed() error {

	wallet.Lock()
	defer wallet.Unlock()

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return errors.New("Entropy of the address raised an error")
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return errors.New("Mnemonic couldn't be created")
	}

	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = seed
	return nil
}

func (wallet *Wallet) createEmptyWallet() (err error) {
	wallet.createSeed()
	_, err = wallet.AddNewAddress()
	return
}

func (wallet *Wallet) updateWallet() {
	gui.InfoUpdate("Wallet", wallet.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}

func (wallet *Wallet) computeChecksum() []byte {

	data, err := helpers.GetJSON(wallet, "Checksum")
	if err != nil {
		panic(err)
	}

	return cryptography.GetChecksum(data)
}

func (wallet *Wallet) Close() {

}
