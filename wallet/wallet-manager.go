package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) GetAddressesCount() int {
	wallet.Lock()
	defer wallet.Unlock()
	return len(wallet.Addresses)
}

func (wallet *Wallet) GetFirstWalletForDevnetGenesisAirdrop() (string, []byte, error) {

	wallet.Lock()
	defer wallet.Unlock()

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

func (wallet *Wallet) DecodeBalanceByPublicKey(publicKey []byte, balance *crypto.ElGamal, token []byte, store, lock bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if lock {

		if store {
			wallet.Lock()
			defer wallet.Unlock()
		} else {
			wallet.RLock()
			defer wallet.RUnlock()
		}

	}

	addr := wallet.addressesMap[string(publicKey)]
	if addr == nil {
		return 0, errors.New("address was not found")
	}

	decoded, err := addr.DecodeBalance(balance, token, store, ctx, statusCallback)
	if err != nil {
		return 0, err
	}

	if store {
		if err := wallet.saveWalletAddress(addr, false); err != nil {
			gui.GUI.Error("error storing balance update", publicKey)
		}
	}

	return decoded, nil
}

func (wallet *Wallet) GetWalletAddressByEncodedAddress(addressEncoded string) (*wallet_address.WalletAddress, error) {

	address, err := addresses.DecodeAddr(addressEncoded)
	if err != nil {
		return nil, err
	}

	wallet.RLock()
	defer wallet.RUnlock()

	out := wallet.addressesMap[string(address.PublicKey)]
	if out == nil {
		return nil, errors.New("address was not found")
	}

	return out, nil
}

func (wallet *Wallet) GetWalletAddressByPublicKey(publicKey []byte) *wallet_address.WalletAddress {

	wallet.RLock()
	defer wallet.RUnlock()

	return wallet.addressesMap[string(publicKey)]
}

func (wallet *Wallet) GetDataForDecodingBalance(publicKey, token []byte) (privateKey []byte, previousValue uint64) {

	wallet.RLock()
	defer wallet.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]
	privateKey = addr.PrivateKey.Key

	if addr.BalancesDecoded[string(token)] != nil {
		previousValue = addr.BalancesDecoded[string(token)].AmountDecoded
	}

	return
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
		wallet.Lock()
		defer wallet.Unlock()
	}
	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	address, err := addresses.NewAddr(config.NETWORK_SELECTED, addresses.SIMPLE_PUBLIC_KEY, adr.PublicKey, nil, 0, nil)
	if err != nil {
		return
	}

	adr.AddressEncoded = address.EncodeAddr()

	if wallet.addressesMap[string(adr.PublicKey)] != nil {
		return errors.New("Address exists")
	}

	wallet.Addresses = append(wallet.Addresses, adr)
	wallet.addressesMap[string(adr.PublicKey)] = adr

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKey)

	wallet.Count += 1
	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKey)

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
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return errors.New("Wallet was not loaded!")
	}

	var addr1, addr2 *addresses.Address
	if addr1, err = adr.PrivateKey.GenerateAddress(false, 0, []byte{}); err != nil {
		return
	}

	if addr2, err = adr.PrivateKey.GenerateAddress(true, 0, []byte{}); err != nil {
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

	wallet.forging.Wallet.AddWallet(adr.GetDelegatedStakePrivateKey(), adr.PublicKey)

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
		Name:           "Addr " + strconv.FormatUint(uint64(wallet.SeedIndex), 10),
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

func (wallet *Wallet) RemoveAddressByIndex(index int, lock bool) (bool, error) {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
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

	wallet.forging.Wallet.RemoveWallet(removing.PublicKey)

	wallet.updateWallet()
	if err := wallet.saveWallet(index, index+1, wallet.Count, false); err != nil {
		return false, err
	}
	globals.MainEvents.BroadcastEvent("wallet/removed", adr)

	return true, nil
}

func (wallet *Wallet) RemoveAddress(encodedAddress string, lock bool) (bool, error) {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	for i, addr := range wallet.Addresses {
		if addr.AddressEncoded == encodedAddress {
			return wallet.RemoveAddressByIndex(i, false)
		}
	}

	return false, nil
}

func (wallet *Wallet) RemoveAddressByWalletAddress(address *wallet_address.WalletAddress, lock bool) (bool, error) {

	if lock {
		wallet.Lock()
		defer wallet.Unlock()
	}

	if !wallet.Loaded {
		return false, errors.New("Wallet was not loaded!")
	}

	for i, addr := range wallet.Addresses {
		if addr == address {
			return wallet.RemoveAddressByIndex(i, false)
		}
	}

	return false, nil
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

func (wallet *Wallet) refreshWalletAccount(acc *account.Account, adr *wallet_address.WalletAddress, lock bool) (err error) {

	if acc == nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	adr.DecodeAccount(acc, true, ctx, func(string) {})

	return
}

func (wallet *Wallet) refreshWalletPlainAccount(plainAcc *plain_account.PlainAccount, adr *wallet_address.WalletAddress, lock bool) (err error) {

	if plainAcc == nil {
		return
	}

	if adr.DelegatedStake != nil && plainAcc.DelegatedStake == nil {
		adr.DelegatedStake = nil

		if adr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByWalletAddress(adr, lock)
			return
		}

		return
	}

	if (adr.DelegatedStake != nil && plainAcc.DelegatedStake != nil && !bytes.Equal(adr.DelegatedStake.PublicKey, plainAcc.DelegatedStake.DelegatedPublicKey)) ||
		(adr.DelegatedStake == nil && plainAcc.DelegatedStake != nil) {

		if adr.PrivateKey == nil {
			_, err = wallet.RemoveAddressByWalletAddress(adr, lock)
			return
		}

		if plainAcc.DelegatedStake != nil {

			lastKnownNonce := uint32(0)
			if adr.DelegatedStake != nil {
				lastKnownNonce = adr.DelegatedStake.LastKnownNonce
			}

			var delegatedStake *wallet_address.WalletAddressDelegatedStake
			if delegatedStake, err = adr.FindDelegatedStake(uint32(plainAcc.Nonce), lastKnownNonce, plainAcc.DelegatedStake.DelegatedPublicKey); err != nil {
				return
			}

			if delegatedStake != nil {
				adr.DelegatedStake = delegatedStake
				wallet.forging.Wallet.AddWallet(adr.DelegatedStake.PrivateKey.Key, adr.PublicKey)
				return wallet.saveWalletAddress(adr, lock)
			}

		}

		adr.DelegatedStake = nil
		wallet.forging.Wallet.AddWallet(nil, adr.PublicKey)

		return wallet.saveWalletAddress(adr, lock)
	}

	return
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

	wallet2 := createWallet(wallet.forging, wallet.mempool, wallet.updateAccounts, wallet.updatePlainAccounts)
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
		wallet.addressesMap[string(adr.PublicKey)] = adr
	}

	return
}

func (wallet *Wallet) GetDelegatesCount() int {
	wallet.RLock()
	defer wallet.RUnlock()

	return wallet.DelegatesCount
}

func (wallet *Wallet) Close() {

}
