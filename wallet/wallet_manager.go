package wallet

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/derivation"
	"pandora-pay/wallet/wallet_address"
	"pandora-pay/wallet/wallet_address/shared_staked"
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

func (wallet *Wallet) GetFirstAddressForDevnetGenesisAirdrop() (string, *shared_staked.WalletAddressSharedStakedAddressExported, error) {

	addr, err := wallet.GetFirstStakedAddress(true)
	if err != nil {
		return "", nil, err
	}

	sharedStakedAddress, err := wallet.exportSharedStakedAddress(addr, "", false)
	if err != nil {
		return "", nil, err
	}

	return addr.AddressEncoded, sharedStakedAddress, nil
}

func (wallet *Wallet) GetWalletAddressByEncodedAddress(addressEncoded string, lock bool) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	return wallet.GetWalletAddressByPublicKeyHash(address.PublicKeyHash, lock), nil
}

func (wallet *Wallet) GetWalletAddressByPublicKeyHashString(publicKeyHashStr string, lock bool) (*wallet_address.WalletAddress, error) {
	publicKeyHash, err := base64.StdEncoding.DecodeString(publicKeyHashStr)
	if err != nil {
		return nil, err
	}
	return wallet.GetWalletAddressByPublicKeyHash(publicKeyHash, lock), nil
}

func (wallet *Wallet) GetWalletAddressByPublicKeyHash(publicKeyHash []byte, lock bool) *wallet_address.WalletAddress {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	return wallet.addressesMap[string(publicKeyHash)].Clone()
}

func (wallet *Wallet) ImportSecretKey(name string, secretKey []byte) (*wallet_address.WalletAddress, error) {

	secret, err := derivation.NewMasterKey(secretKey)
	if err != nil {
		return nil, err
	}

	secretRaw := secret.RawSeed()

	privateKey, err := secret.Derive(derivation.FirstHardenedIndex)
	if err != nil {
		return nil, err
	}

	privKey, err := privateKey.GetPrivateKey()
	if err != nil {
		return nil, err
	}

	privKeyObj, err := addresses.NewPrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	addr := &wallet_address.WalletAddress{
		Name:       name,
		SecretKey:  secretRaw[:],
		PrivateKey: privKeyObj,
		SeedIndex:  0,
		IsImported: true,
		IsMine:     true,
	}

	if err = wallet.AddAddress(addr, true, false, false, true); err != nil {
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

	address, err := addresses.CreateAddr(addr.PublicKeyHash, nil, 0, nil)
	if err != nil {
		return
	}

	addr.AddressEncoded = address.EncodeAddr()

	if wallet.addressesMap[string(addr.PublicKeyHash)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, addr)
	wallet.addressesMap[string(addr.PublicKeyHash)] = addr

	wallet.forging.Wallet.AddWallet(addr.PublicKeyHash, addr.SharedStaked, false, nil, 0)

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
	publicKeyHash := cryptography.GetPublicKeyHash(publicKey)

	addr.AddressEncoded = addr1.EncodeAddr()
	addr.PublicKey = publicKey
	addr.PublicKeyHash = publicKeyHash

	if addr.PrivateKey != nil {
		if addr.SharedStaked, err = addr.DeriveSharedStaked(0); err != nil {
			return
		}
	}

	if wallet.addressesMap[string(addr.PublicKeyHash)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, addr)
	wallet.addressesMap[string(addr.PublicKeyHash)] = addr

	wallet.Count += 1

	if incrementSeedIndex {
		wallet.SeedIndex += 1
	}
	if incrementImportedCountIndex {
		addr.Name = "Imported Address " + strconv.Itoa(wallet.CountImportedIndex)
		wallet.CountImportedIndex += 1
	}

	if err = wallet.forging.Wallet.AddWallet(addr.PublicKeyHash, addr.SharedStaked, false, nil, 0); err != nil {
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

	seedExtend := &addresses.SeedExtended{}
	if err := seedExtend.Deserialize(wallet.Seed); err != nil {
		return nil, nil, err
	}

	masterKey, err := derivation.NewMasterKey(seedExtend.Key)
	if err != nil {
		return nil, nil, err
	}

	secret, err := masterKey.Derive(derivation.FirstHardenedIndex + seedIndex)
	if err != nil {
		return nil, nil, err
	}

	privateKey, err := secret.Derive(derivation.FirstHardenedIndex)
	if err != nil {
		return nil, nil, err
	}

	seed := secret.RawSeed()

	privKey, err := privateKey.GetPrivateKey()
	if err != nil {
		return nil, nil, err
	}

	return seed[:], privKey[:], nil
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

	privKey, err := addresses.NewPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

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
	delete(wallet.addressesMap, string(adr.PublicKeyHash))

	wallet.Count -= 1

	wallet.forging.Wallet.RemoveWallet(removing.PublicKeyHash, false, nil, 0)

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

	return wallet.RemoveAddressByPublicKeyHash(addr.PublicKeyHash, lock)
}

func (wallet *Wallet) RemoveAddressByPublicKeyHash(publicKeyHash []byte, lock bool) (bool, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	for i, addr := range wallet.Addresses {
		if bytes.Equal(addr.PublicKeyHash, publicKeyHash) {
			return wallet.RemoveAddressByIndex(i, false)
		}
	}

	return false, nil
}

func (wallet *Wallet) RenameAddressByPublicKeyHash(publicKeyHash []byte, newName string, lock bool) (bool, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	addr := wallet.GetWalletAddressByPublicKeyHash(publicKeyHash, false)
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

func (wallet *Wallet) GetAddressSecretKey(index int) ([]byte, error) { //32 byte

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
		addr.SeedIndex = 0
		addr.IsImported = true
	}
	addr.IsMine = true

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
		wallet.addressesMap[string(adr.PublicKeyHash)] = adr
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
