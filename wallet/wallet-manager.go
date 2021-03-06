package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) GetAddressesCount() int {
	wallet.Lock()
	defer wallet.Unlock()
	return len(wallet.Addresses)
}

func (wallet *Wallet) GetFirstWalletForDevnetGenesisAirdrop() ([]byte, []byte, error) {

	wallet.Lock()
	defer wallet.Unlock()

	if len(wallet.Addresses) == 0 || !wallet.Loaded {
		return nil, nil, errors.New("Wallet is empty")
	}

	addr := wallet.Addresses[0]
	delegatedStake, err := addr.DeriveDelegatedStake(0)
	if err != nil {
		return nil, nil, err
	}

	return addr.PublicKeyHash, delegatedStake.PublicKeyHash, nil
}

func (wallet *Wallet) GetWalletAddressByEncodedAddress(addressEncoded string) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	wallet.RLock()
	defer wallet.RUnlock()

	out := wallet.addressesMap[string(address.PublicKeyHash)]
	if out == nil {
		return nil, errors.New("address was not found")
	}

	return out, nil
}

func (wallet *Wallet) ImportPrivateKey(name string, privateKey []byte) (*wallet_address.WalletAddress, error) {

	if len(privateKey) != 32 {
		return nil, errors.New("Invalid PrivateKey length")
	}

	addr := &wallet_address.WalletAddress{
		Name:           name,
		PrivateKey:     &addresses.PrivateKey{Key: privateKey},
		SeedIndex:      1,
		DelegatedStake: nil,
		IsMine:         true,
	}

	if err := wallet.AddAddress(addr, true, false, false); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) AddAddress(adr *wallet_address.WalletAddress, lock bool, incrementSeedIndex bool, incrementCountIndex bool) (err error) {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	var addr1 *addresses.Address
	if addr1, err = adr.PrivateKey.GenerateAddress(true, 0, []byte{}); err != nil {
		return
	}

	var publicKey, publicKeyHash []byte
	publicKey, publicKeyHash, err = adr.PrivateKey.GeneratePairs()

	adr.AddressEncoded = addr1.EncodeAddr()
	adr.PublicKey = publicKey
	adr.PublicKeyHash = publicKeyHash

	if wallet.addressesMap[string(adr.PublicKeyHash)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, adr)
	wallet.addressesMap[string(adr.PublicKeyHash)] = adr

	wallet.Count += 1

	if incrementSeedIndex {
		wallet.SeedIndex += 1
	}
	if incrementCountIndex {
		adr.Name = "Imported Address " + strconv.Itoa(wallet.CountIndex)
		wallet.CountIndex += 1
	}

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKeyHash)
	wallet.mempool.Wallet.AddWallet(adr.PublicKeyHash)

	wallet.updateWallet()
	gui.GUI.Info("wallet.saveWallet", len(wallet.Addresses))
	if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("wallet/added", adr)

	return

}

func (wallet *Wallet) GeneratePrivateKey(seedIndex uint32, lock bool) ([]byte, error) {
	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return nil, errors.New("Wallet was not loaded!")
	}

	masterKey, err := bip32.NewMasterKey(wallet.Seed)
	if err != nil {
		return nil, err
	}

	key, err := masterKey.NewChildKey(seedIndex)
	if err != nil {
		return nil, err
	}

	return key.Key, nil
}

func (wallet *Wallet) AddNewAddress(lock bool) (*wallet_address.WalletAddress, error) {

	//avoid generating the same address twice
	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return nil, errors.New("Wallet was not loaded!")
	}

	key, err := wallet.GeneratePrivateKey(wallet.SeedIndex, false)
	if err != nil {
		return nil, err
	}

	adr := &wallet_address.WalletAddress{
		Name:           "Addr " + strconv.Itoa(wallet.Count),
		PrivateKey:     &addresses.PrivateKey{Key: key},
		SeedIndex:      wallet.SeedIndex,
		DelegatedStake: nil,
		IsMine:         true,
	}

	if err = wallet.AddAddress(adr, false, true, false); err != nil {
		return nil, err
	}

	return adr, nil
}

/**
In case encodedAddress is provided, it will search for it
*/
func (wallet *Wallet) RemoveAddress(index int, encodedAddress string) (bool, error) {

	wallet.Lock()
	defer wallet.Unlock()

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	if encodedAddress != "" {
		for i, addr := range wallet.Addresses {
			if addr.AddressEncoded == encodedAddress {
				index = i
				break
			}
		}
	}

	if index < 0 || index > len(wallet.Addresses) {
		return false, errors.New("Invalid Address Index")
	}

	adr := wallet.Addresses[index]

	removing := wallet.Addresses[index]
	wallet.Addresses = append(wallet.Addresses[:index], wallet.Addresses[index+1:]...)
	delete(wallet.addressesMap, string(adr.PublicKeyHash))

	wallet.Count -= 1

	wallet.forging.Wallet.RemoveWallet(removing.PublicKeyHash)
	wallet.mempool.Wallet.RemoveWallet(removing.PublicKeyHash)

	wallet.updateWallet()
	if err := wallet.saveWallet(index, len(wallet.Addresses), wallet.Count, false); err != nil {
		return false, err
	}
	globals.MainEvents.BroadcastEvent("wallet/removed", adr)

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

func (wallet *Wallet) createSeed(lock bool) error {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

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
	wallet.Lock()
	defer wallet.Unlock()

	wallet.setLoaded(true)
	if err = wallet.createSeed(false); err != nil {
		return
	}
	if _, err = wallet.AddNewAddress(false); err != nil {
		return
	}

	return
}

func (wallet *Wallet) updateWallet() {
	gui.GUI.InfoUpdate("Wallet", wallet.Encryption.Encrypted.String())
	gui.GUI.InfoUpdate("Wallet Addrs", strconv.Itoa(wallet.Count))
}

//wallet must be locked before
//acc read only
func (wallet *Wallet) refreshWallet(acc *account.Account, adr *wallet_address.WalletAddress, lock bool) (err error) {

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
					wallet.forging.Wallet.AddWallet(adr.DelegatedStake.PrivateKey.Key, adr.PublicKeyHash)
					return wallet.saveWalletAddress(adr, lock)
				}

			}

		}

		adr.DelegatedStake = nil
		wallet.forging.Wallet.AddWallet(nil, adr.PublicKeyHash)
		return wallet.saveWalletAddress(adr, lock)
	}

	return
}

func (wallet *Wallet) updateAccountsChanges() {

	var err error
	updateAccountsCn := wallet.updateAccounts.AddListener()
	defer wallet.updateAccounts.RemoveChannel(updateAccountsCn)

	for {
		accsData, ok := <-updateAccountsCn
		if !ok {
			return
		}

		accs := accsData.(*accounts.Accounts)

		wallet.Lock()
		for k, v := range accs.HashMap.Committed {
			if wallet.addressesMap[k] != nil {

				if v.Stored == "update" {
					acc := v.Element.(*account.Account)
					if err = wallet.refreshWallet(acc, wallet.addressesMap[k], false); err != nil {
						return
					}
				} else if v.Stored == "delete" {
					if err = wallet.refreshWallet(nil, wallet.addressesMap[k], false); err != nil {
						return
					}
				}

			}
		}
		wallet.Unlock()
	}

}

func (wallet *Wallet) ImportWalletAddressJSON(data []byte) (*wallet_address.WalletAddress, error) {

	adr := &wallet_address.WalletAddress{}

	if err := json.Unmarshal(data, adr); err != nil {
		return nil, errors.New("Error unmarshaling wallet")
	}

	wallet.RLock()
	defer wallet.RUnlock()

	isMine := false
	if wallet.SeedIndex != 0 {
		key, err := wallet.GeneratePrivateKey(adr.SeedIndex, false)
		if err == nil && key != nil {
			isMine = true
		}
	}

	if !isMine {
		adr.IsMine = false
		adr.SeedIndex = 0
	}

	if err := wallet.AddAddress(adr, false, false, isMine); err != nil {
		return nil, err
	}

	return adr, nil
}

func (wallet *Wallet) ImportWalletJSON(data []byte) (err error) {

	wallet2 := createWallet(wallet.forging, wallet.mempool, wallet.updateAccounts)
	if err = json.Unmarshal(data, wallet2); err != nil {
		return errors.New("Error unmarshaling wallet")
	}

	wallet.RLock()
	defer wallet.RUnlock()

	if err = json.Unmarshal(data, wallet); err != nil {
		return errors.New("Error unmarshaling wallet 2")
	}

	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	for _, adr := range wallet.Addresses {
		wallet.addressesMap[string(adr.PublicKeyHash)] = adr
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
