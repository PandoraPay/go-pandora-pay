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
		if bytes.Equal(addr.PublicKey[:], address.PublicKey) || bytes.Equal(addr.PublicKeyHash[:], address.PublicKey) {
			out = addr
			return
		}
	}

	panic("address was not found")
}

func (wallet *Wallet) addNewAddress() {

	//avoid generating the same address twice
	wallet.Lock()
	defer wallet.Unlock()

	masterKey, _ := bip32.NewMasterKey(wallet.Seed[:])

	key, err := masterKey.NewChildKey(wallet.SeedIndex)
	if err != nil {
		panic("Couldn't derivate the marker key")
	}

	privateKey := addresses.PrivateKey{Key: *helpers.Byte32(key.Key)}
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
}

func (wallet *Wallet) removeAddress(index int) {

	wallet.Lock()
	defer wallet.Unlock()

	if index < 0 || index > len(wallet.Addresses) {
		panic("Invalid Address Index")
	}

	removing := wallet.Addresses[index]

	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	go wallet.forging.Wallet.RemoveWallet(removing.PublicKey)

	wallet.updateWallet()
	wallet.saveWallet(index, wallet.Count, wallet.Count)
}

func (wallet *Wallet) showPrivateKey(index int) [32]byte {

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
	wallet.Seed = *helpers.Byte32(seed)

}

func (wallet *Wallet) createEmptyWallet() {
	wallet.createSeed()
	wallet.addNewAddress()
}

func (wallet *Wallet) updateWallet() {
	gui.InfoUpdate("Wallet", wallet.Encrypted.String())
	gui.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}

func (wallet *Wallet) computeChecksum() (checksum [4]byte) {

	data, err := helpers.GetJSON(wallet, "Checksum")
	if err != nil {
		panic(err)
	}

	out := cryptography.RIPEMD(data)[0:helpers.ChecksumSize]
	copy(checksum[:], out[:])

	return
}
