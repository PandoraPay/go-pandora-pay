package wallet

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"pandora-pay/addresses"
	"pandora-pay/config"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/wallet/wallet_address"
	"strconv"
)

func (wallet *Wallet) GetAddressesCount() int {
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()
	return len(wallet.Addresses)
}

func (wallet *Wallet) GetFirstAddressForDevnetGenesisAirdrop() (string, []byte, error) {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	if len(wallet.Addresses) == 0 || !wallet.Loaded {
		return "", nil, errors.New("Wallet is empty")
	}

	addr := wallet.Addresses[0]
	delegatedStake, err := addr.DeriveDelegatedStake(0)
	if err != nil {
		return "", nil, err
	}

	return addr.AddressRegistrationEncoded, delegatedStake.PublicKey, nil
}

//you should not lock it before
func (wallet *Wallet) DecryptBalanceByPublicKey(publicKey []byte, balance, asset []byte, useNewPreviousValue bool, newPreviousValue uint64, store, lock bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if len(balance) == 0 {
		return 0, errors.New("Encrypted Balance is nil")
	}

	if !lock {
		return 0, errors.New("You shouldn't lock the wallet before as it will lock wallet functionality for some time")
	}

	if lock && store {
		wallet.Lock.Lock()
	} else if lock && !store {
		wallet.Lock.RLock()
	}

	addr := wallet.addressesMap[string(publicKey)]
	if addr == nil {
		return 0, errors.New("address was not found")
	}

	priv := &addresses.PrivateKey{helpers.CloneBytes(addr.PrivateKey.Key)}

	var previousValue uint64
	if !useNewPreviousValue {
		if found := addr.DecryptedBalances[base64.StdEncoding.EncodeToString(asset)]; found != nil {
			previousValue = found.Amount
		}
	} else {
		previousValue = newPreviousValue
	}

	if lock && store {
		wallet.Lock.Unlock()
	} else if lock && !store {
		wallet.Lock.RUnlock()
	}

	balancePoint, err := new(crypto.ElGamal).Deserialize(balance)
	if err != nil {
		return 0, err
	}

	decrypted, err := priv.DecryptBalance(balancePoint, previousValue, ctx, statusCallback)
	if err != nil {
		return 0, err
	}

	if store {
		if lock {
			wallet.Lock.Lock()
			defer wallet.Lock.Unlock()
		}
		if addr = wallet.addressesMap[string(publicKey)]; addr == nil {
			return 0, errors.New("address for storing the new decrypted value was not found")
		}

		addr.UpdateDecryptedBalance(decrypted, balance, asset)

		if err := wallet.saveWalletAddress(addr, false); err != nil {
			gui.GUI.Error("error storing balance update", publicKey)
		}
	}

	return decrypted, nil
}

func (wallet *Wallet) UpdatePreviousDecryptedBalanceValueByPublicKey(publicKey []byte, newDecodedBalance uint64, balanceEncrypted, asset []byte) error {
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	addr := wallet.addressesMap[string(publicKey)]
	if addr == nil {
		return errors.New("address was not found")
	}

	addr.UpdateDecryptedBalance(newDecodedBalance, balanceEncrypted, asset)

	return wallet.saveWalletAddress(addr, false)
}

func (wallet *Wallet) GetWalletAddressByEncodedAddress(addressEncoded string, lock bool) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	return wallet.GetWalletAddressByPublicKey(address.PublicKey, lock), nil
}

func (wallet *Wallet) GetWalletAddressByPublicKeyString(publicKeyStr string, lock bool) (*wallet_address.WalletAddress, error) {
	publicKey, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, err
	}
	return wallet.GetWalletAddressByPublicKey(publicKey, lock), nil
}

func (wallet *Wallet) GetWalletAddressByPublicKey(publicKey []byte, lock bool) *wallet_address.WalletAddress {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	return wallet.addressesMap[string(publicKey)].Clone()
}

func (wallet *Wallet) TryDecryptBalance(publicKey, asset, balance []byte) (uint64, bool, error) {

	balancePoint, err := new(crypto.ElGamal).Deserialize(balance)
	if err != nil {
		return 0, false, err
	}

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]

	if addr.DecryptedBalances[base64.StdEncoding.EncodeToString(asset)] == nil {
		return 0, false, nil
	}

	previousValue := addr.DecryptedBalances[base64.StdEncoding.EncodeToString(asset)].Amount

	ok := addr.PrivateKey.TryDecryptBalance(balancePoint, previousValue)
	if ok {
		return previousValue, true, nil
	}

	return 0, false, nil
}

func (wallet *Wallet) ImportPrivateKey(name string, privateKey []byte) (*wallet_address.WalletAddress, error) {

	if len(privateKey) != 32 {
		return nil, errors.New("Invalid PrivateKey length")
	}

	priv := &addresses.PrivateKey{Key: privateKey}
	reg, err := priv.GetRegistration()
	if err != nil {
		return nil, err
	}

	addr := &wallet_address.WalletAddress{
		Name:           name,
		PrivateKey:     priv,
		Registration:   reg,
		SeedIndex:      1,
		DelegatedStake: nil,
		IsMine:         true,
	}

	if err := wallet.AddAddress(addr, true, false, false); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) AddDelegateStakeAddress(adr *wallet_address.WalletAddress, lock bool) (err error) {
	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}
	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	if wallet.Count > config_nodes.DELEGATES_MAXIMUM {
		return errors.New("DELEGATES_MAXIMUM exceeded")
	}

	address, err := addresses.NewAddr(config.NETWORK_SELECTED, addresses.SIMPLE_PUBLIC_KEY, adr.PublicKey, nil, nil, 0, nil)
	if err != nil {
		return
	}

	adr.AddressEncoded = address.EncodeAddr()

	if wallet.addressesMap[string(adr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, adr)
	wallet.addressesMap[string(adr.PublicKey)] = adr

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKey, false, nil)

	wallet.Count += 1

	wallet.updateWallet()

	if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("wallet/added", adr)

	return
}

func (wallet *Wallet) AddAddress(adr *wallet_address.WalletAddress, lock bool, incrementSeedIndex bool, incrementImportedCountIndex bool) (err error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	var addr1, addr2 *addresses.Address
	if addr1, err = adr.PrivateKey.GenerateAddress(false, nil, 0, nil); err != nil {
		return
	}

	if addr2, err = adr.PrivateKey.GenerateAddress(true, nil, 0, nil); err != nil {
		return
	}

	publicKey := adr.PrivateKey.GeneratePublicKey()

	if adr.DecryptedBalances == nil {
		adr.DecryptedBalances = make(map[string]*wallet_address.WalletAddressDecryptedBalance)
	}
	adr.AddressEncoded = addr1.EncodeAddr()
	adr.AddressRegistrationEncoded = addr2.EncodeAddr()
	adr.PublicKey = publicKey

	if wallet.addressesMap[string(adr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, adr)
	wallet.addressesMap[string(adr.PublicKey)] = adr

	wallet.Count += 1

	if incrementSeedIndex {
		wallet.SeedIndex += 1
	}
	if incrementImportedCountIndex {
		adr.Name = "Imported Address " + strconv.Itoa(wallet.CountImportedIndex)
		wallet.CountImportedIndex += 1
	}

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKey, false, nil)

	wallet.updateWallet()

	if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("wallet/added", adr)

	return

}

func (wallet *Wallet) GeneratePrivateKey(seedIndex uint32, lock bool) ([]byte, error) {
	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
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
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return nil, errors.New("Wallet was not loaded!")
	}

	privateKey, err := wallet.GeneratePrivateKey(wallet.SeedIndex, false)
	if err != nil {
		return nil, err
	}

	key := &addresses.PrivateKey{Key: privateKey}
	reg, err := key.GetRegistration()
	if err != nil {
		return nil, err
	}

	adr := &wallet_address.WalletAddress{
		Name:           "Addr_" + strconv.FormatUint(uint64(wallet.SeedIndex), 10),
		PrivateKey:     key,
		Registration:   reg,
		SeedIndex:      wallet.SeedIndex,
		DelegatedStake: nil,
		IsMine:         true,
	}

	if err = wallet.AddAddress(adr, false, true, false); err != nil {
		return nil, err
	}

	return adr, nil
}

func (wallet *Wallet) DeriveDelegatedStakeByPublicKey(addressPublicKey []byte, nonce uint64) ([]byte, []byte, error) {
	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.GetWalletAddressByPublicKey(addressPublicKey, false)
	if addr == nil {
		return nil, nil, errors.New("Wallet was not found")
	}

	walletAddressDelegatedStake, err := addr.DeriveDelegatedStake(uint32(nonce))
	if err != nil {
		return nil, nil, err
	}
	return walletAddressDelegatedStake.PublicKey, walletAddressDelegatedStake.PrivateKey.Key, nil
}

func (wallet *Wallet) RemoveAddressByIndex(index int, lock bool) (bool, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	if index < 0 || index > len(wallet.Addresses) {
		return false, errors.New("Invalid Address Index")
	}

	adr := wallet.Addresses[index]

	removing := wallet.Addresses[index]

	wallet.Addresses[index] = wallet.Addresses[len(wallet.Addresses)-1]
	wallet.Addresses = wallet.Addresses[:len(wallet.Addresses)-1]
	delete(wallet.addressesMap, string(adr.PublicKey))

	wallet.Count -= 1

	wallet.forging.Wallet.RemoveWallet(removing.PublicKey, false, nil)

	wallet.updateWallet()
	if err := wallet.saveWallet(index, index+1, wallet.Count, false); err != nil {
		return false, err
	}
	globals.MainEvents.BroadcastEvent("wallet/removed", adr)

	return true, nil
}

func (wallet *Wallet) RemoveAddress(encodedAddress string, lock bool) (bool, error) {

	addr, err := addresses.DecodeAddr(encodedAddress)
	if err != nil {
		return false, err
	}

	return wallet.RemoveAddressByPublicKey(addr.PublicKey, lock)
}

func (wallet *Wallet) RemoveAddressByPublicKey(publicKey []byte, lock bool) (bool, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	for i, addr := range wallet.Addresses {
		if bytes.Equal(addr.PublicKey, publicKey) {
			return wallet.RemoveAddressByIndex(i, false)
		}
	}

	return false, nil
}

func (wallet *Wallet) RenameAddressByPublicKey(publicKey []byte, newName string, lock bool) (bool, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	addr := wallet.GetWalletAddressByPublicKey(publicKey, false)
	if addr == nil {
		return false, nil
	}

	addr.Name = newName

	return true, wallet.saveWalletAddress(addr, false)
}

func (wallet *Wallet) GetWalletAddress(index int, lock bool) (*wallet_address.WalletAddress, error) {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].Clone(), nil
}

func (wallet *Wallet) GetPrivateKey(index int) ([]byte, error) { //32 byte

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key, nil
}

func (wallet *Wallet) ImportWalletAddressJSON(data []byte) (*wallet_address.WalletAddress, error) {

	addr := &wallet_address.WalletAddress{}
	if err := json.Unmarshal(data, addr); err != nil {
		return nil, errors.New("Error unmarshaling wallet")
	}

	if addr.PrivateKey == nil {
		return nil, errors.New("Private Key is missing")
	}

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	isMine := false
	if wallet.SeedIndex != 0 {
		key, err := wallet.GeneratePrivateKey(addr.SeedIndex, false)
		if err == nil && key != nil && bytes.Equal(key, addr.PrivateKey.Key) {
			isMine = true
		}
	}

	if !isMine {
		addr.IsMine = false
		addr.SeedIndex = 0
	}

	if err := wallet.AddAddress(addr, false, false, isMine); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) ImportWalletJSON(data []byte) (err error) {

	wallet2 := createWallet(wallet.forging, wallet.mempool, wallet.updateAccounts, wallet.updatePlainAccounts)
	if err = json.Unmarshal(data, wallet2); err != nil {
		return errors.New("Error unmarshaling wallet")
	}

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	if err = json.Unmarshal(data, wallet); err != nil {
		return errors.New("Error unmarshaling wallet 2")
	}

	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	for _, adr := range wallet.Addresses {
		wallet.addressesMap[string(adr.PublicKey)] = adr
	}

	return
}

func (wallet *Wallet) GetDelegatesCount() int {
	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	return wallet.DelegatesCount
}

func (wallet *Wallet) Close() {

}
