package wallet

import (
	"bytes"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
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

func (wallet *Wallet) GetWalletAddressByAddress(addressEncoded string) (out *wallet_address.WalletAddress, err error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	wallet.RLock()
	defer wallet.RUnlock()

	out = wallet.AddressesMap[string(address.PublicKeyHash)]
	if out == nil {
		err = errors.New("address was not found")
	}

	return
}

func (wallet *Wallet) ImportPrivateKey(name string, privateKey []byte) (adr *wallet_address.WalletAddress, err error) {

	if len(privateKey) != 32 {
		errors.New("Invalid PrivateKey length")
	}

	wallet.RLock()
	defer wallet.RUnlock()

	adr = &wallet_address.WalletAddress{
		Name:           name,
		PrivateKey:     &addresses.PrivateKey{Key: privateKey},
		SeedIndex:      0,
		DelegatedStake: nil,
		IsMine:         true,
	}

	err = wallet.AddAddress(adr, false)

	return
}

func (wallet *Wallet) AddAddress(adr *wallet_address.WalletAddress, lock bool) (err error) {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if adr.Address, err = adr.PrivateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		return
	}

	adr.AddressEncoded = adr.Address.EncodeAddr()

	wallet.Addresses = append(wallet.Addresses, adr)
	wallet.AddressesMap[string(adr.Address.PublicKeyHash)] = adr

	wallet.Count += 1
	wallet.SeedIndex += 1

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.GetPublicKeyHash())
	wallet.mempool.Wallet.AddWallet(adr.GetPublicKeyHash())

	wallet.updateWallet()
	if err = wallet.saveWallet(wallet.Count-1, wallet.Count, -1); err != nil {
		return
	}

	return

}

func (wallet *Wallet) AddNewAddress() (adr *wallet_address.WalletAddress, err error) {

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

	adr = &wallet_address.WalletAddress{
		Name:           "Addr " + strconv.Itoa(wallet.Count),
		PrivateKey:     &addresses.PrivateKey{Key: key.Key},
		SeedIndex:      wallet.SeedIndex,
		DelegatedStake: nil,
		IsMine:         true,
	}

	wallet.Count += 1
	wallet.SeedIndex += 1

	if err = wallet.AddAddress(adr, false); err != nil {
		wallet.Count -= 1
		wallet.SeedIndex -= 1
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

	adr := wallet.Addresses[index]

	removing := wallet.Addresses[index]
	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	delete(wallet.AddressesMap, string(adr.Address.PublicKeyHash))

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

//wallet must be locked before
func (wallet *Wallet) refreshWallet(acc *account.Account, adr *wallet_address.WalletAddress) (err error) {

	if acc == nil {
		return
	}

	if adr.DelegatedStake != nil && acc.DelegatedStake == nil {
		adr.DelegatedStake = nil
		return
	}

	if (adr.DelegatedStake != nil && acc.DelegatedStake != nil && !bytes.Equal(adr.DelegatedStake.PublicKeyHash, acc.DelegatedStake.DelegatedPublicKeyHash)) ||
		(adr.DelegatedStake == nil && acc.DelegatedStake != nil) {

		if adr.IsMine {

			if acc.DelegatedStake != nil {

				lastKnownNonce := uint32(0)
				if adr.DelegatedStake != nil {
					lastKnownNonce = adr.DelegatedStake.LastKnownNonce
				}

				var delegatedStake *wallet_address.WalletAddressDelegatedStake
				if delegatedStake, err = adr.FindDelegatedStake(uint32(acc.Nonce), lastKnownNonce, acc.DelegatedStake.DelegatedPublicKeyHash); err != nil {
					return
				}

				if delegatedStake != nil {
					adr.DelegatedStake = delegatedStake
					wallet.forging.Wallet.AddWallet(adr.DelegatedStake.PrivateKey.Key, adr.Address.PublicKeyHash)
					return wallet.saveWalletAddress(adr)
				}

			}

		}

		adr.DelegatedStake = nil
		wallet.forging.Wallet.AddWallet(nil, adr.Address.PublicKeyHash)
		return wallet.saveWalletAddress(adr)
	}

	return
}

func (wallet *Wallet) UpdateAccountsChanges(accs *accounts.Accounts) (err error) {

	wallet.Lock()
	defer wallet.Unlock()

	for k, v := range accs.HashMap.Committed {
		if wallet.AddressesMap[k] != nil {

			if v.Commit == "update" {
				acc := new(account.Account)
				if err = acc.Deserialize(helpers.NewBufferReader(v.Data)); err != nil {
					return
				}
				if err = wallet.refreshWallet(acc, wallet.AddressesMap[k]); err != nil {
					return
				}
			} else if v.Commit == "delete" {
				if err = wallet.refreshWallet(nil, wallet.AddressesMap[k]); err != nil {
					return
				}
			}

		}
	}

	return
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
