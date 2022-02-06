package wallet

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
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

func (wallet *Wallet) GetFirstWalletForDevnetGenesisAirdrop() (string, []byte, error) {

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
func (wallet *Wallet) DecodeBalanceByPublicKey(publicKey []byte, balance *crypto.ElGamal, asset []byte, store, lock bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

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

	previousValue := uint64(0)
	if found := addr.BalancesDecoded[hex.EncodeToString(asset)]; found != nil {
		previousValue = found.AmountDecoded
	}

	if lock && store {
		wallet.Lock.Unlock()
	} else if lock && !store {
		wallet.Lock.RUnlock()
	}

	decoded, err := priv.DecodeBalance(balance, previousValue, ctx, statusCallback)
	if err != nil {
		return 0, err
	}

	if store {
		if lock {
			wallet.Lock.Lock()
			defer wallet.Lock.Unlock()
		}
		if addr = wallet.addressesMap[string(publicKey)]; addr == nil {
			return 0, errors.New("address for storing the new decoded value was not found")
		}

		addr.UpdatePreviousValue(decoded, asset)

		if err := wallet.saveWalletAddress(addr, false); err != nil {
			gui.GUI.Error("error storing balance update", publicKey)
		}
	}

	return decoded, nil
}

func (wallet *Wallet) UpdatePreviousValueByPublicKey(publicKey []byte, newPreviousValue uint64, asset []byte) error {
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	addr := wallet.addressesMap[string(publicKey)]
	if addr == nil {
		return errors.New("address was not found")
	}

	addr.UpdatePreviousValue(newPreviousValue, asset)

	return wallet.saveWalletAddress(addr, false)
}

func (wallet *Wallet) GetWalletAddressByEncodedAddress(addressEncoded string) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	out := wallet.addressesMap[string(address.PublicKey)]
	if out == nil {
		return nil, errors.New("address was not found")
	}

	return out, nil
}

func (wallet *Wallet) GetWalletAddressByPublicKeyHex(publicKeyHex string) (*wallet_address.WalletAddress, error) {
	publicKey, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return nil, err
	}
	return wallet.GetWalletAddressByPublicKey(publicKey, false), nil
}

func (wallet *Wallet) GetWalletAddressByPublicKey(publicKey []byte, lock bool) *wallet_address.WalletAddress {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	return wallet.addressesMap[string(publicKey)]
}

func (wallet *Wallet) DecodeBalanceIfMatchesPreviousValue(publicKey, asset, balanceEncoded []byte) (uint64, bool, error) {

	balance, err := new(crypto.ElGamal).Deserialize(balanceEncoded)
	if err != nil {
		return 0, false, err
	}

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]

	if addr.BalancesDecoded[hex.EncodeToString(asset)] == nil {
		return 0, false, nil
	}

	previousValue := addr.BalancesDecoded[hex.EncodeToString(asset)].AmountDecoded

	ok := addr.PrivateKey.CheckMatchBalanceDecoded(balance, previousValue)
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
	gui.GUI.Info("wallet.saveWallet", len(wallet.Addresses))
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

	if adr.BalancesDecoded == nil {
		adr.BalancesDecoded = make(map[string]*wallet_address.WalletAddressBalanceDecoded)
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
	gui.GUI.Info("wallet.saveWallet", len(wallet.Addresses))
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
	return wallet.Addresses[index], nil
}

func (wallet *Wallet) ShowPrivateKey(index int) ([]byte, error) { //32 byte

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	if index < 0 || index > len(wallet.Addresses) {
		return nil, errors.New("Invalid Address Index")
	}
	return wallet.Addresses[index].PrivateKey.Key, nil
}

func (wallet *Wallet) createSeed(lock bool) error {

	if lock {
		wallet.Lock.Lock()
		defer wallet.Lock.Unlock()
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
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

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
	gui.GUI.InfoUpdate("Wallet Addrs", fmt.Sprintf("%d  %s", wallet.Count, wallet.Encryption.Encrypted))
}

func (wallet *Wallet) refreshWalletPlainAccount(plainAcc *plain_account.PlainAccount, adr *wallet_address.WalletAddress, lock bool) (err error) {

	if plainAcc == nil {
		return
	}

	if adr.DelegatedStake != nil && !plainAcc.DelegatedStake.HasDelegatedStake() {
		adr.DelegatedStake = nil

		if adr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByPublicKey(adr.PublicKey, lock)
			return
		}

		return
	}

	if (adr.DelegatedStake != nil && plainAcc.DelegatedStake.HasDelegatedStake() && !bytes.Equal(adr.DelegatedStake.PublicKey, plainAcc.DelegatedStake.DelegatedStakePublicKey)) ||
		(adr.DelegatedStake == nil && plainAcc.DelegatedStake.HasDelegatedStake()) {

		if adr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByPublicKey(adr.PublicKey, lock)
			return
		}

		if plainAcc.DelegatedStake.HasDelegatedStake() {

			if plainAcc.DelegatedStake.DelegatedStakeFee < config_nodes.DELEGATOR_FEE {
				_, err = wallet.RemoveAddressByPublicKey(adr.PublicKey, lock)
				return
			}

			lastKnownNonce := uint32(0)
			if adr.DelegatedStake != nil {
				lastKnownNonce = adr.DelegatedStake.LastKnownNonce
			}

			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = adr.FindDelegatedStake(uint32(plainAcc.Nonce), lastKnownNonce, plainAcc.DelegatedStake.DelegatedStakePublicKey); err != nil {
				_, err = wallet.RemoveAddressByPublicKey(adr.PublicKey, lock)
				return
			}

			if delegatedStake != nil {
				adr.DelegatedStake = delegatedStake
				wallet.forging.Wallet.AddWallet(adr.DelegatedStake.PrivateKey.Key, adr.PublicKey, true, plainAcc)
				return wallet.saveWalletAddress(adr, lock)
			}

		}

		adr.DelegatedStake = nil
		wallet.forging.Wallet.RemoveWallet(adr.PublicKey, true, plainAcc)

		return wallet.saveWalletAddress(adr, lock)
	}

	return
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
