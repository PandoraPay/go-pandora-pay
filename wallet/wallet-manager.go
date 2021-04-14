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
	wallet.AddressesMap[string(walletAddress.Address.PublicKeyHash)] = walletAddress

	wallet.Count += 1
	wallet.SeedIndex += 1

	wallet.forging.Wallet.AddWallet(walletAddress.GetDelegatedStakePrivateKey(), walletAddress.GetPublicKeyHash())
	wallet.mempool.Wallet.AddWallet(walletAddress.GetPublicKeyHash())

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

	addr := wallet.Addresses[index]

	removing := wallet.Addresses[index]
	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	delete(wallet.AddressesMap, string(addr.Address.PublicKeyHash))

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
func (wallet *Wallet) refreshWallet(acc *account.Account, addr *wallet_address.WalletAddress) (err error) {

	if acc == nil {
		return
	}

	if addr.DelegatedStake != nil && acc.DelegatedStake == nil {
		addr.DelegatedStake = nil
		return
	}

	if (addr.DelegatedStake != nil && acc.DelegatedStake != nil && !bytes.Equal(addr.DelegatedStake.PublicKeyHash, acc.DelegatedStake.DelegatedPublicKeyHash)) ||
		(addr.DelegatedStake == nil && acc.DelegatedStake != nil) {

		if addr.IsMine {

			if acc.DelegatedStake != nil {

				var delegatedStake *wallet_address.WalletAddressDelegatedStake
				if delegatedStake, err = addr.FindDelegatedStake(uint32(acc.Nonce), acc.DelegatedStake.DelegatedPublicKeyHash); err != nil {
					return
				}

				if delegatedStake != nil {
					addr.DelegatedStake = delegatedStake
					wallet.forging.Wallet.AddWallet(addr.DelegatedStake.PrivateKey.Key, addr.Address.PublicKeyHash)
					return
				}
			}

		}

		addr.DelegatedStake = nil
		wallet.forging.Wallet.AddWallet(nil, addr.Address.PublicKeyHash)
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
				if err = acc.Deserialize(v.Data); err != nil {
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
