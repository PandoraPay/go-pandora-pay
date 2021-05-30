package wallet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config/globals"
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

func (wallet *Wallet) saveWallet(start, end, deleteIndex int, lock bool) error {

	if lock {
		wallet.RLock()
		defer wallet.RUnlock()
	}

	return store.StoreWallet.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return
		}

		marshal, err := helpers.GetJSON(wallet, "addresses", "addressesMap")
		if err != nil {
			return
		}

		if err = writer.Put([]byte("wallet"), marshal); err != nil {
			return
		}

		for i := start; i < end; i++ {
			gui.GUI.Log("Saving WALLET", i)
			if marshal, err = json.Marshal(wallet.Addresses[i]); err != nil {
				return
			}
			if err = writer.Put([]byte("wallet-address-"+strconv.Itoa(i)), marshal); err != nil {
				return
			}
		}

		if deleteIndex != -1 {
			if err = writer.Delete([]byte("wallet-address-" + strconv.Itoa(deleteIndex))); err != nil {
				return
			}
		}

		if err = writer.Put([]byte("saved"), []byte{1}); err != nil {
			return
		}

		return
	})
}

func (wallet *Wallet) loadWallet() error {

	return store.StoreWallet.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {

			gui.GUI.Log("Wallet Loading... ")

			var unmarshal []byte

			unmarshal = reader.Get([]byte("wallet"))

			if err = json.Unmarshal(unmarshal, &wallet); err != nil {
				return
			}

			wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
			wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)

			for i := 0; i < wallet.Count; i++ {
				unmarshal := reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))

				newWalletAddress := &wallet_address.WalletAddress{}
				if err = json.Unmarshal(unmarshal, newWalletAddress); err != nil {
					return
				}
				wallet.Addresses = append(wallet.Addresses, newWalletAddress)
				wallet.addressesMap[string(newWalletAddress.PublicKeyHash)] = newWalletAddress

				wallet.forging.Wallet.AddWallet(newWalletAddress.GetDelegatedStakePrivateKey(), newWalletAddress.PublicKeyHash)
				wallet.mempool.Wallet.AddWallet(newWalletAddress.PublicKeyHash)

			}

			wallet.updateWallet()
			globals.MainEvents.BroadcastEvent("wallet/loaded", wallet.Count)
			gui.GUI.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}

func (wallet *Wallet) ReadWallet() error {

	wallet.Lock()
	defer wallet.Unlock()

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		chainHeight, _ := binary.Uvarint(reader.Get([]byte("chainHeight")))

		accs := accounts.NewAccounts(reader)
		for _, adr := range wallet.Addresses {

			var acc *account.Account
			if acc, err = accs.GetAccount(adr.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if err = wallet.refreshWallet(acc, adr, false); err != nil {
				return
			}

		}

		return
	})
}
