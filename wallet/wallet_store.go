package wallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
	"strconv"
)

func (wallet *Wallet) saveWalletAddress(adr *wallet_address.WalletAddress, lock bool) error {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	for i, adr2 := range wallet.Addresses {
		if adr2 == adr {
			return wallet.saveWallet(i, i+1, -1, false)
		}
	}

	return nil
}

func (wallet *Wallet) saveWalletEntire(lock bool) error {
	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}
	return wallet.saveWallet(0, wallet.Count, -1, false)
}

func (wallet *Wallet) saveWallet(start, end, deleteIndex int, lock bool) error {

	if lock {
		wallet.Lock.RLock()
		defer wallet.Lock.RUnlock()
	}

	start = generics.Max(0, start)
	end = generics.Min(end, len(wallet.Addresses))

	if !wallet.Loaded {
		return errors.New("Can't save your wallet because your stored wallet on the drive was not successfully loaded")
	}

	return store.StoreWallet.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		var marshal []byte

		writer.Put("saved", []byte{0})

		if marshal, err = helpers.GetMarshalledDataExcept(wallet.Encryption); err != nil {
			return
		}
		writer.Put("encryption", marshal)

		if marshal, err = helpers.GetMarshalledDataExcept(wallet, "addresses", "encryption"); err != nil {
			return
		}
		if marshal, err = wallet.Encryption.encryptData(marshal); err != nil {
			return
		}

		writer.Put("wallet", marshal)

		for i := start; i < end; i++ {
			if marshal, err = msgpack.Marshal(wallet.Addresses[i]); err != nil {
				return
			}
			if marshal, err = wallet.Encryption.encryptData(marshal); err != nil {
				return
			}
			writer.Put("wallet-address-"+strconv.Itoa(i), marshal)
		}
		if deleteIndex != -1 {
			writer.Delete("wallet-address-" + strconv.Itoa(deleteIndex))
		}

		writer.Put("saved", []byte{1})
		return
	})
}

func (wallet *Wallet) loadWallet(password string, first bool) error {
	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	if wallet.Loaded {
		return errors.New("Wallet was already loaded!")
	}

	wallet.clearWallet()

	return store.StoreWallet.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		saved := reader.Get("saved") //safe only internal
		if saved == nil {
			return errors.New("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {

			gui.GUI.Log("Wallet Loading... ")

			var unmarshal []byte

			unmarshal = reader.Get("encryption")
			if unmarshal == nil {
				return errors.New("encryption data was not found")
			}
			if err = msgpack.Unmarshal(unmarshal, wallet.Encryption); err != nil {
				return
			}

			if wallet.Encryption.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
				if password == "" {
					return nil
				}
				wallet.Encryption.password = password
				if err = wallet.Encryption.createEncryptionCipher(); err != nil {
					return
				}
			}

			if unmarshal, err = wallet.Encryption.decryptData(reader.Get("wallet")); err != nil {
				return
			}
			if err = msgpack.Unmarshal(unmarshal, wallet); err != nil {
				return
			}

			wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
			wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)

			for i := 0; i < wallet.Count; i++ {

				if unmarshal, err = wallet.Encryption.decryptData(reader.Get("wallet-address-" + strconv.Itoa(i))); err != nil {
					return
				}

				newWalletAddress := &wallet_address.WalletAddress{}
				if err = msgpack.Unmarshal(unmarshal, newWalletAddress); err != nil {
					return
				}

				if newWalletAddress.PrivateKey != nil {
					if !bytes.Equal(newWalletAddress.PrivateKey.GeneratePublicKey(), newWalletAddress.PublicKey) {
						return errors.New("Public Keys are not matching!")
					}
				}

				wallet.Addresses = append(wallet.Addresses, newWalletAddress)
				wallet.addressesMap[string(newWalletAddress.PublicKey)] = newWalletAddress

			}

			wallet.setLoaded(true)
			if !first {
				if err = wallet.walletLoaded(); err != nil {
					return
				}
			}

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}

func (wallet *Wallet) walletLoaded() error {

	go wallet.InitForgingWallet()

	wallet.updateWallet()
	globals.MainEvents.BroadcastEvent("wallet/loaded", wallet.Count)
	gui.GUI.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

	return nil
}

func (wallet *Wallet) StartWallet() error {

	wallet.Lock.Lock()
	defer wallet.Lock.Unlock()

	return wallet.walletLoaded()
}

func (wallet *Wallet) InitForgingWallet() (err error) {

	if !config_forging.FORGING_ENABLED {
		return nil
	}

	for _, addr := range wallet.Addresses {
		if err = wallet.forging.Wallet.AddWallet(addr.PublicKey, addr.SharedStaked, false, nil, nil, 0); err != nil {
			return
		}
	}

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)
		var accs *accounts.Accounts
		if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
			return
		}

		for _, addr := range wallet.Addresses {

			var acc *account.Account
			var reg *registration.Registration

			if acc, err = accs.GetAccount(addr.PublicKey); err != nil {
				return
			}
			if reg, err = dataStorage.Regs.GetRegistration(addr.PublicKey); err != nil {
				return
			}

			if err = wallet.refreshWalletAccount(acc, reg, chainHeight, addr); err != nil {
				return
			}
		}

		return
	})
}
