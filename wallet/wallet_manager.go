package wallet

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/config/config_nodes"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
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
	}

	var found *wallet_address.WalletAddress
	for _, addr := range wallet.Addresses {
		if addr.Staked {
			found = addr
			break
		}
	}

	if lock {
		wallet.Lock.RUnlock()
	}
	if found != nil {
		return found, nil
	}

	return wallet.AddNewAddress(true, "", true, true, true)
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

	return addr.AddressRegistrationEncoded, sharedStakedAddress, nil
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

func (wallet *Wallet) ImportSecretKey(name string, secret []byte, staked, spendRequired bool) (*wallet_address.WalletAddress, error) {

	secretPrivateKeyExtended := &addresses.PrivateKeyExtended{}
	if err := secretPrivateKeyExtended.Deserialize(secret); err != nil {
		return nil, err
	}

	secretChild, err := bip32.Deserialize(secretPrivateKeyExtended.Key)
	if err != nil {
		return nil, err
	}

	start := uint32(0)
	if secretPrivateKeyExtended.Version == addresses.SIMPLE_PRIVATE_KEY_WIF { //hardened
		start = bip32.FirstHardenedChild
	}

	privKey, err := secretChild.NewChildKey(start + 0)
	if err != nil {
		return nil, err
	}

	spendPrivKey, err := secretChild.NewChildKey(start + 1)
	if err != nil {
		return nil, err
	}

	privateKey, err := addresses.NewPrivateKey(privKey.Key)
	if err != nil {
		return nil, err
	}

	spendPrivateKey, err := addresses.NewPrivateKey(spendPrivKey.Key)
	if err != nil {
		return nil, err
	}

	addr := &wallet_address.WalletAddress{
		Name:            name,
		SecretKey:       secret,
		PrivateKey:      privateKey,
		SeedIndex:       1,
		SpendPrivateKey: spendPrivateKey,
		IsMine:          true,
	}

	if err = wallet.AddAddress(addr, staked, spendRequired, true, false, false, true); err != nil {
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

	address, err := addresses.CreateAddr(addr.PublicKey, addr.Staked, addr.SpendPublicKey, nil, nil, 0, nil)
	if err != nil {
		return
	}

	addr.AddressEncoded = address.EncodeAddr()

	if wallet.addressesMap[string(addr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, addr)
	wallet.addressesMap[string(addr.PublicKey)] = addr

	wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, false, nil, nil, 0)

	wallet.Count += 1

	wallet.updateWallet()

	if err = wallet.saveWallet(len(wallet.Addresses)-1, len(wallet.Addresses), -1, false); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("wallet/added", addr)

	return
}

func (wallet *Wallet) AddAddress(addr *wallet_address.WalletAddress, staked, spendRequired, lock bool, incrementSeedIndex, incrementImportedCountIndex, save bool) (err error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	if addr.SpendPrivateKey != nil {
		addr.SpendPublicKey = addr.SpendPrivateKey.GeneratePublicKey()
	}

	var spendPublicKey []byte
	if spendRequired {
		if len(addr.SpendPublicKey) != cryptography.PublicKeySize {
			return errors.New("Spend Public Key is missing")
		}
		spendPublicKey = addr.SpendPublicKey
	}

	var addr1, addr2 *addresses.Address

	if addr1, err = addr.PrivateKey.GenerateAddress(staked, spendPublicKey, false, nil, 0, nil); err != nil {
		return
	}
	if addr2, err = addr.PrivateKey.GenerateAddress(staked, spendPublicKey, true, nil, 0, nil); err != nil {
		return
	}

	publicKey := addr.PrivateKey.GeneratePublicKey()

	addr.Staked = staked
	addr.SpendRequired = spendRequired
	addr.AddressEncoded = addr1.EncodeAddr()
	addr.AddressRegistrationEncoded = addr2.EncodeAddr()
	addr.Registration = addr2.Registration
	addr.PublicKey = publicKey

	if addr.PrivateKey != nil {
		if addr.SharedStaked, err = addr.DeriveSharedStaked(); err != nil {
			return
		}
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

	if err = wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, false, nil, nil, 0); err != nil {
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

func (wallet *Wallet) GenerateKeys(seedIndex uint32, lock bool) ([]byte, []byte, []byte, error) {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	if !wallet.Loaded {
		return nil, nil, nil, errors.New("Wallet was not loaded!")
	}

	seedExtend := &addresses.PrivateKeyExtended{}
	if err := seedExtend.Deserialize(wallet.Seed); err != nil {
		return nil, nil, nil, err
	}

	masterKey, err := bip32.NewMasterKey(seedExtend.Key)
	if err != nil {
		return nil, nil, nil, err
	}

	var index uint32
	if seedExtend.Version == addresses.SIMPLE_PRIVATE_KEY_WIF {
		index = bip32.FirstHardenedChild
	}

	secret, err := masterKey.NewChildKey(index + seedIndex)
	if err != nil {
		return nil, nil, nil, err
	}

	key2, err := secret.NewChildKey(index + 0)
	if err != nil {
		return nil, nil, nil, err
	}

	key3, err := secret.NewChildKey(index + 1)
	if err != nil {
		return nil, nil, nil, err
	}

	secretSerialized, err := secret.Serialize()
	if err != nil {
		return nil, nil, nil, err
	}

	return secretSerialized, key2.Key, key3.Key, nil
}

func (wallet *Wallet) AddNewAddress(lock bool, name string, staked, spendRequired, save bool) (*wallet_address.WalletAddress, error) {

	//avoid generating the same address twice
	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
	}

	version := wallet_address.VERSION_NORMAL

	if !wallet.Loaded {
		return nil, errors.New("Wallet was not loaded!")
	}

	secret, privateKey, spendPrivateKey, err := wallet.GenerateKeys(wallet.SeedIndex, false)
	if err != nil {
		return nil, err
	}

	privKey, err := addresses.NewPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	spendPrivKey, err := addresses.NewPrivateKey(spendPrivateKey)
	if err != nil {
		return nil, err
	}

	if name == "" {
		name = "Addr_" + strconv.FormatUint(uint64(wallet.SeedIndex), 10)
	}

	addr := &wallet_address.WalletAddress{
		Version:         version,
		Name:            name,
		SecretKey:       secret,
		PrivateKey:      privKey,
		SpendPrivateKey: spendPrivKey,
		SeedIndex:       wallet.SeedIndex,
		IsMine:          true,
	}

	if err = wallet.AddAddress(addr, staked, spendRequired, false, true, false, save); err != nil {
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

	wallet.forging.Wallet.RemoveWallet(removing.PublicKey, false, nil, nil, 0)

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
		key, _, _, err := wallet.GenerateKeys(addr.SeedIndex, false)
		if err == nil && key != nil && bytes.Equal(key, addr.PrivateKey.Key) {
			isMine = true
		}
	}

	if !isMine {
		addr.IsMine = false
		addr.SeedIndex = 0
	}

	if err := wallet.AddAddress(addr, addr.Staked, addr.SpendRequired, false, false, isMine, true); err != nil {
		return nil, err
	}

	return addr, nil
}

func (wallet *Wallet) DecryptBalance(addr *wallet_address.WalletAddress, encryptedBalance, asset []byte, useNewPreviousValue bool, newPreviousValue uint64, store bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if len(encryptedBalance) == 0 {
		return 0, errors.New("Encrypted Balance is nil")
	}

	return wallet.addressBalanceDecryptor.DecryptBalance("wallet", addr.PublicKey, addr.PrivateKey.Key, encryptedBalance, asset, useNewPreviousValue, newPreviousValue, store, ctx, statusCallback)
}

func (wallet *Wallet) DecryptBalanceByPublicKey(publicKey []byte, encryptedBalance, asset []byte, useNewPreviousValue bool, newPreviousValue uint64, store, lock bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

	addr := wallet.GetWalletAddressByPublicKey(publicKey, lock)
	if addr == nil {
		return 0, errors.New("address was not found")
	}

	return wallet.DecryptBalance(addr, encryptedBalance, asset, useNewPreviousValue, newPreviousValue, store, ctx, statusCallback)
}

func (wallet *Wallet) TryDecryptBalanceByPublicKey(publicKey []byte, encryptedBalance []byte, lock bool, matchValue uint64) (bool, error) {

	if len(encryptedBalance) == 0 {
		return false, errors.New("Encrypted Balance is nil")
	}

	addr := wallet.GetWalletAddressByPublicKey(publicKey, lock)
	if addr == nil {
		return false, errors.New("address was not found")
	}

	return wallet.TryDecryptBalance(addr, encryptedBalance, matchValue)
}

func (wallet *Wallet) TryDecryptBalance(addr *wallet_address.WalletAddress, encryptedBalance []byte, matchValue uint64) (bool, error) {
	balance, err := new(crypto.ElGamal).Deserialize(encryptedBalance)
	if err != nil {
		return false, err
	}

	return addr.PrivateKey.TryDecryptBalance(balance, matchValue), nil
}

func (wallet *Wallet) ImportWalletJSON(data []byte) (err error) {

	wallet2 := createWallet(wallet.forging, wallet.mempool, wallet.addressBalanceDecryptor, wallet.updateNewChainUpdate)
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
