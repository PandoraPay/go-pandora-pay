package wallet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/encryption"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	wallet_address "pandora-pay/wallet/address"
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

	if !wallet.loaded {
		return errors.New("Can't save your wallet because your stored wallet on the drive was not successfully loaded")
	}

	return store.StoreWallet.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		var marshal []byte

		if err = writer.Put("saved", []byte{0}); err != nil {
			return
		}

		if marshal, err = helpers.GetJSON(wallet.Encryption); err != nil {
			return
		}
		if err = writer.Put("encryption", marshal); err != nil {
			return
		}

		if marshal, err = helpers.GetJSON(wallet, "addresses", "encryption"); err != nil {
			return
		}
		if wallet.Encryption.Encrypted == ENCRYPTED_VERSION_ENCRYPTION_ARGON2 {
			if marshal, err = wallet.Encryption.encryptionCipher.Encrypt(marshal); err != nil {
				return
			}
		}
		if err = writer.Put("wallet", marshal); err != nil {
			return
		}

		for i := start; i < end; i++ {
			if marshal, err = json.Marshal(wallet.Addresses[i]); err != nil {
				return
			}
			if wallet.Encryption.Encrypted == ENCRYPTED_VERSION_ENCRYPTION_ARGON2 {
				if marshal, err = wallet.Encryption.encryptionCipher.Encrypt(marshal); err != nil {
					return
				}
			}
			if err = writer.Put("wallet-address-"+strconv.Itoa(i), marshal); err != nil {
				return
			}
		}

		if deleteIndex != -1 {
			if err = writer.Delete("wallet-address-" + strconv.Itoa(deleteIndex)); err != nil {
				return
			}
		}

		if err = writer.Put("saved", []byte{1}); err != nil {
			return
		}

		return
	})
}

func (wallet *Wallet) loadWallet(password string) error {
	wallet.Lock()
	defer wallet.Unlock()

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
			if err = json.Unmarshal(unmarshal, &wallet.Encryption); err != nil {
				return
			}

			if wallet.Encryption.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
				if wallet.Encryption.encryptionCipher, err = encryption.CreateEncryptionCipher(password, wallet.Encryption.Salt); err != nil {
					return
				}
			}

			unmarshal = reader.Get("wallet")
			if wallet.Encryption.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
				if unmarshal, err = wallet.Encryption.encryptionCipher.Decrypt(unmarshal); err != nil {
					return
				}
			}
			if err = json.Unmarshal(unmarshal, &wallet); err != nil {
				return
			}

			wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
			wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)

			for i := 0; i < wallet.Count; i++ {

				unmarshal := reader.Get("wallet-address-" + strconv.Itoa(i))
				if wallet.Encryption.Encrypted != ENCRYPTED_VERSION_PLAIN_TEXT {
					if unmarshal, err = wallet.Encryption.encryptionCipher.Decrypt(unmarshal); err != nil {
						return
					}
				}

				newWalletAddress := &wallet_address.WalletAddress{}
				if err = json.Unmarshal(unmarshal, newWalletAddress); err != nil {
					return
				}
				wallet.Addresses = append(wallet.Addresses, newWalletAddress)
				wallet.addressesMap[string(newWalletAddress.PublicKeyHash)] = newWalletAddress

			}

			for _, addr := range wallet.Addresses {
				wallet.forging.Wallet.AddWallet(addr.GetDelegatedStakePrivateKey(), addr.PublicKeyHash)
				wallet.mempool.Wallet.AddWallet(addr.PublicKeyHash)
			}

			wallet.loaded = true
			wallet.updateWallet()
			globals.MainEvents.BroadcastEvent("wallet/loaded", wallet.Count)
			gui.GUI.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}

func (wallet *Wallet) StartWallet() error {

	wallet.Lock()
	defer wallet.Unlock()

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))

		accs := accounts.NewAccounts(reader)
		for _, adr := range wallet.Addresses {

			var acc, acc2 *account.Account
			if acc2, err = accs.GetAccount(adr.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if acc2 != nil { //let's clone it
				acc = &account.Account{}
				if err = acc.Deserialize(helpers.NewBufferReader(helpers.CloneBytes(acc2.SerializeToBytes()))); err != nil {
					return
				}
			}

			if err = wallet.refreshWallet(acc, adr, false); err != nil {
				return
			}

		}

		return
	})
}
