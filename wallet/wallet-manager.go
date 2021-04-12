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
	"pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) GetFirstWalletForDevnetGenesisAirdrop() (adr *wallet_address.WalletAddress, delegatedPublicKeyhash []byte, err error) {
	wallet.Lock()
	defer wallet.Unlock()

	adr = wallet.Addresses[0]
	delegatedStake, err := adr.DeriveDelegatedStake(0)
	if err != nil {
		return
	}

	return adr, delegatedStake.PublicKeyHash, nil
}

func (wallet *Wallet) GetWalletAddressByAddress(addressEncoded string) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	wallet.RLock()
	defer wallet.RUnlock()

	for _, addr := range wallet.Addresses {
		if bytes.Equal(addr.GetPublicKeyHash(), address.PublicKeyHash) {
			return addr, nil
		}
	}

	return nil, errors.New("address was not found")
}

func (wallet *Wallet) AddNewAddress() (walletAddress *wallet_address.WalletAddress, err error) {

	//avoid generating the same address twice
	wallet.Lock()
	defer wallet.Unlock()

	masterKey, err := bip32.NewMasterKey(wallet.Seed)
	if err != nil {
		return
	}

	key, err := masterKey.NewChildKey(wallet.SeedIndex)
	if err != nil {
		return
	}

	walletAddress = &wallet_address.WalletAddress{
		Name:           "Addr " + strconv.Itoa(wallet.Count),
		PrivateKey:     &addresses.PrivateKey{Key: key.Key},
		SeedIndex:      wallet.SeedIndex,
		DelegatedStake: nil,
		IsMine:         true,
	}

	if walletAddress.Address, err = walletAddress.PrivateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		return
	}

	wallet.Addresses = append(wallet.Addresses, walletAddress)
	wallet.Count += 1
	wallet.SeedIndex += 1

	go wallet.forging.Wallet.AddWallet(walletAddress.GetDelegatedStakePrivateKey(), walletAddress.GetPublicKeyHash())
	go wallet.mempool.Wallet.AddWallet(walletAddress.GetPublicKeyHash())

	wallet.updateWallet()
	if err = wallet.saveWallet(wallet.Count-1, wallet.Count, -1); err != nil {
		return
	}

	return
}

func (wallet *Wallet) RemoveAddress(index int) (out bool, err error) {

	wallet.Lock()
	defer wallet.Unlock()

	if index < 0 || index > len(wallet.Addresses) {
		return false, errors.New("Invalid Address Index")
	}

	removing := wallet.Addresses[index]

	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	wallet.Count -= 1

	wallet.forging.Wallet.RemoveWallet(removing.GetPublicKeyHash())
	wallet.mempool.Wallet.RemoveWallet(removing.GetPublicKeyHash())

	wallet.updateWallet()
	if err = wallet.saveWallet(index, wallet.Count, wallet.Count); err != nil {
		return
	}

	return true, nil
}

func (wallet *Wallet) GetWalletAddress(index int) (*wallet_address.WalletAddress, error) {
	wallet.RLock()
	defer wallet.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index], nil
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
	if err = wallet.createSeed(); err != nil {
		return
	}
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
