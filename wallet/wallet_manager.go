package wallet

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/config"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/globals"
	"pandora-pay/wallet/wallet_address"
	"strconv"
)

func (wallet *Wallet) GetAddressesCount() int {
	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()
	return len(wallet.Addresses)
}

func (wallet *Wallet) GetRandomAddress() *wallet_address.WalletAddress {
	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()
	index := rand.Intn(len(wallet.Addresses))
	return wallet.Addresses[index].Clone()
}

func (wallet *Wallet) GetFirstStakedAddress(lock bool) (*wallet_address.WalletAddress, error) {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	return wallet.Addresses[0].Clone(), nil
}

func (wallet *Wallet) GetFirstAddressForDevnetGenesisAirdrop() (string, error) {

	addr, err := wallet.GetFirstStakedAddress(true)
	if err != nil {
		return "", err
	}

	return addr.AddressEncoded, nil
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

func (wallet *Wallet) ImportSecretKey(name string, secretKey []byte) (*wallet_address.WalletAddress, error) {

	secretChild, err := bip32.Deserialize(secretKey)
	if err != nil {
		return nil, err
	}

	privKey, err := secretChild.NewChildKey(0)
	if err != nil {
		return nil, err
	}

	privateKey := &addresses.PrivateKey{Key: privKey.Key}

	addr := &wallet_address.WalletAddress{
		Name:       name,
		SecretKey:  secretKey,
		PrivateKey: privateKey,
		SeedIndex:  1,
		IsMine:     true,
	}

	if err := wallet.AddAddress(addr, true, false, false, true); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) AddSharedStakedAddress(addr *wallet_address.WalletAddress, lock bool) (err error) {

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

	address, err := addresses.NewAddr(config.NETWORK_SELECTED, addresses.SIMPLE_PUBLIC_KEY, addr.PublicKey, nil, 0, nil)
	if err != nil {
		return
	}

	addr.AddressEncoded = address.EncodeAddr()

	if wallet.addressesMap[string(addr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, addr)
	wallet.addressesMap[string(addr.PublicKey)] = addr

	wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, false, nil, 0)

	wallet.Count += 1

	wallet.updateWallet()

	if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("wallet/added", addr)

	return
}

func (wallet *Wallet) AddAddress(addr *wallet_address.WalletAddress, lock bool, incrementSeedIndex, incrementImportedCountIndex, save bool) (err error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	var addr1 *addresses.Address

	if addr1, err = addr.PrivateKey.GenerateAddress(nil, 0, nil); err != nil {
		return
	}

	publicKey := addr.PrivateKey.GeneratePublicKey()

	addr.AddressEncoded = addr1.EncodeAddr()
	addr.PublicKey = publicKey

	if addr.PrivateKey != nil {
		addr.SharedStaked = &wallet_address.WalletAddressSharedStaked{addr.PrivateKey, addr.PublicKey}
	}

	if wallet.addressesMap[string(addr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, addr)
	wallet.addressesMap[string(addr.PublicKey)] = addr

	wallet.Count += 1

	if incrementSeedIndex {
		wallet.SeedIndex += 1
	}
	if incrementImportedCountIndex {
		addr.Name = "Imported Address " + strconv.Itoa(wallet.CountImportedIndex)
		wallet.CountImportedIndex += 1
	}

	if err = wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, false, nil, 0); err != nil {
		return
	}

	if save {
		wallet.updateWallet()

		if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
			return
		}
		globals.MainEvents.BroadcastEvent("wallet/added", addr)
	}

	return

}

func (wallet *Wallet) GenerateKeys(seedIndex uint32, lock bool) ([]byte, []byte, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return nil, nil, errors.New("Wallet was not loaded!")
	}

	masterKey, err := bip32.NewMasterKey(wallet.Seed)
	if err != nil {
		return nil, nil, err
	}

	secret, err := masterKey.NewChildKey(seedIndex)
	if err != nil {
		return nil, nil, err
	}

	key2, err := secret.NewChildKey(0)
	if err != nil {
		return nil, nil, err
	}

	secretSerialized, err := secret.Serialize()
	if err != nil {
		return nil, nil, err
	}

	return secretSerialized, key2.Key, nil
}

func (wallet *Wallet) AddNewAddress(lock bool, name string, save bool) (*wallet_address.WalletAddress, error) {

	//avoid generating the same address twice
	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	version := wallet_address.VERSION_NORMAL

	if !wallet.Loaded {
		return nil, errors.New("Wallet was not loaded!")
	}

	secret, privateKey, err := wallet.GenerateKeys(wallet.SeedIndex, false)
	if err != nil {
		return nil, err
	}

	privKey := &addresses.PrivateKey{Key: privateKey}

	if name == "" {
		name = "Addr_" + strconv.FormatUint(uint64(wallet.SeedIndex), 10)
	}

	addr := &wallet_address.WalletAddress{
		Version:    version,
		Name:       name,
		SecretKey:  secret,
		PrivateKey: privKey,
		SeedIndex:  wallet.SeedIndex,
		IsMine:     true,
	}

	if err = wallet.AddAddress(addr, false, true, false, save); err != nil {
		return nil, err
	}

	return addr.Clone(), nil
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

	wallet.forging.Wallet.RemoveWallet(removing.PublicKey, false, nil, 0)

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

	if index < 0 || index >= len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].Clone(), nil
}

func (wallet *Wallet) GetSecretKey(index int) ([]byte, error) { //32 byte

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].SecretKey, nil
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
		key, _, err := wallet.GenerateKeys(addr.SeedIndex, false)
		if err == nil && key != nil && bytes.Equal(key, addr.PrivateKey.Key) {
			isMine = true
		}
	}

	if !isMine {
		addr.IsMine = false
		addr.SeedIndex = 0
	}

	if err := wallet.AddAddress(addr, false, false, isMine, true); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) ImportWalletJSON(data []byte) (err error) {

	wallet2 := createWallet(wallet.forging, wallet.mempool, wallet.updateNewChainUpdate)
	if err = json.Unmarshal(data, wallet2); err != nil {
		return errors.New("Error unmarshaling wallet")
	}

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	wallet.clearWallet()
	if err = json.Unmarshal(data, wallet); err != nil {
		return errors.New("Error unmarshaling wallet 2")
	}

	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	for _, adr := range wallet.Addresses {
		wallet.addressesMap[string(adr.PublicKey)] = adr
	}
	wallet.setLoaded(true)

	globals.MainEvents.BroadcastEvent("wallet/loaded", wallet.Count)

	return wallet.saveWalletEntire(false)
}

func (wallet *Wallet) GetDelegatesCount() int {
	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	return wallet.DelegatesCount
}

func (wallet *Wallet) Close() {

}
