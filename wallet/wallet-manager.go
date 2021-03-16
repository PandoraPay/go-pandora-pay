package wallet

import (
	"bytes"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
)

func (wallet *Wallet) GetWalletAddressByAddress(addressEncoded string) (out *WalletAddress) {

	address := addresses.DecodeAddr(addressEncoded)

	for _, addr := range wallet.Addresses {
		if bytes.Equal(addr.PublicKeyHash, address.PublicKeyHash) {
			out = addr
			return
		}
	}

	panic("address was not found")
}

func (wallet *Wallet) AddNewAddress() *WalletAddress {

	//avoid generating the same address twice
	wallet.Lock()
	defer wallet.Unlock()

	masterKey, _ := bip32.NewMasterKey(wallet.Seed)

	key, err := masterKey.NewChildKey(wallet.SeedIndex)
	if err != nil {
		panic("Couldn't derivate the marker key")
	}

	privateKey := addresses.PrivateKey{Key: key.Key}
	publicKey := privateKey.GeneratePublicKey()
	address := privateKey.GenerateAddress(true, 0, []byte{})
	addressEncoded := address.EncodeAddr()
	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)

	walletAddress := WalletAddress{
		"Addr " + strconv.Itoa(wallet.Count),
		&privateKey,
		publicKey,
		publicKeyHash,
		addressEncoded,
		address,
		wallet.SeedIndex,
	}

	wallet.Addresses = append(wallet.Addresses, &walletAddress)
	wallet.Count += 1
	wallet.SeedIndex += 1

	go wallet.forging.Wallet.AddWallet(publicKey, privateKey.Key, publicKeyHash)

	wallet.updateWallet()
	wallet.saveWallet(wallet.Count-1, wallet.Count, -1)

	return &walletAddress
}

func (wallet *Wallet) RemoveAddress(index int) bool {

	wallet.Lock()
	defer wallet.Unlock()

	if index < 0 || index > len(wallet.Addresses) {
		panic("Invalid Address Index")
	}

	removing := wallet.Addresses[index]

	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	go wallet.forging.Wallet.RemoveWallet(removing.PublicKeyHash)

	wallet.updateWallet()
	wallet.saveWallet(index, wallet.Count, wallet.Count)
	return true
}

func (wallet *Wallet) ShowPrivateKey(index int) []byte { //32 byte

	wallet.RLock()
	defer wallet.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		panic("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key
}

func (wallet *Wallet) createSeed() {

	wallet.Lock()
	defer wallet.Unlock()

	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		panic("Entropy of the address raised an error")
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		panic("Mnemonic couldn't be created")
	}

	wallet.Mnemonic = mnemonic

	// Generate a Bip32 HD wallet for the mnemonic and a user supplied password
	seed := bip39.NewSeed(mnemonic, "SEED Secret Passphrase")
	wallet.Seed = seed

}

func (wallet *Wallet) createEmptyWallet() {
	wallet.createSeed()
	wallet.AddNewAddress()
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
