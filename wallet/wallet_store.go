package wallet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
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

	for i, adr2 := range wallet.Addresses {
		if adr2 == adr {
			return wallet.saveWallet(i, i+1, -1, lock)
		}
	}

	return nil
}

func (wallet *Wallet) saveWalletEntire(lock bool) error {
	if lock {
		wallet.RLock()
		defer wallet.RUnlock()
	}
	return wallet.saveWallet(0, wallet.Count, -1, false)
}

func (wallet *Wallet) saveWallet(start, end, deleteIndex int, lock bool) error {

	if lock {
		wallet.RLock()
		defer wallet.RUnlock()
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
	wallet.Lock()
	defer wallet.Unlock()

	if wallet.Loaded {
		return errors.New("Wallet was already loaded!")
	}

	wallet.clearWallet()

	return store.StoreWallet.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		saved := reader.Get("saved")
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

				if !bytes.Equal(newWalletAddress.PrivateKey.GeneratePublicKey(), newWalletAddress.PublicKey) {
					return errors.New("Public Keys are not matching!")
				}

				wallet.Addresses = append(wallet.Addresses, newWalletAddress)
				wallet.addressesMap[string(newWalletAddress.PublicKey)] = newWalletAddress

			}

			wallet.setLoaded(true)
			if !first {
				wallet.walletLoaded()
			}

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}

func (wallet *Wallet) walletLoaded() {

	for _, addr := range wallet.Addresses {
		wallet.forging.Wallet.AddWallet(addr.GetDelegatedStakePrivateKey(), addr.PublicKey, false, nil)
	}

	wallet.updateWallet()
	globals.MainEvents.BroadcastEvent("wallet/loaded", wallet.Count)
	gui.GUI.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

}

func (wallet *Wallet) StartWallet() error {

	wallet.Lock()
	defer wallet.Unlock()

	wallet.walletLoaded()

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		//accsCollection := accounts.NewAccountsCollection(reader)
		//accs, err := accsCollection.GetMap(config_coins.NATIVE_ASSET_FULL)
		//if err != nil {
		//	return
		//}

		//for _, adr := range wallet.Addresses {
		//	var acc *account.Account
		//	if acc, err = accs.GetAccount(adr.PublicKey); err != nil {
		//		return
		//	}
		//
		//	if err = wallet.refreshWalletAccount(acc, adr, false); err != nil {
		//		return
		//	}
		//}

		plainAccs := plain_accounts.NewPlainAccounts(reader)
		for _, adr := range wallet.Addresses {
			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = plainAccs.GetPlainAccount(adr.PublicKey, chainHeight); err != nil {
				return
			}

			if err = wallet.refreshWalletPlainAccount(plainAcc, adr, false); err != nil {
				return
			}
		}

		return
	})
}
