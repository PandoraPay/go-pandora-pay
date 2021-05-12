package wallet

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	wallet_address "pandora-pay/wallet/address"
	"strconv"
)

func (wallet *Wallet) saveWalletAddress(adr *wallet_address.WalletAddress) error {

	for i, adr2 := range wallet.Addresses {
		if adr2 == adr {
			return wallet.saveWallet(i, i+1, -1)
		}
	}

	return nil
}

func (wallet *Wallet) saveWallet(start, end, deleteIndex int) error {
	return store.StoreWallet.DB.Update(func(boltTx *bolt.Tx) (err error) {

		writer := boltTx.Bucket([]byte("Wallet"))

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return
		}

		marshal, err := helpers.GetJSON(wallet, "Addresses", "AddressesMap")
		if err != nil {
			return
		}

		if err = writer.Put([]byte("wallet"), marshal); err != nil {
			return
		}

		for i := start; i < end; i++ {
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

	return store.StoreWallet.DB.View(func(boltTx *bolt.Tx) (err error) {

		reader := boltTx.Bucket([]byte("Wallet"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {

			gui.Log("Wallet Loading... ")

			var unmarshal []byte

			unmarshal = reader.Get([]byte("wallet"))

			if err = json.Unmarshal(unmarshal, &wallet); err != nil {
				return
			}

			for i := 0; i < wallet.Count; i++ {
				unmarshal := reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))

				newWalletAddress := &wallet_address.WalletAddress{}
				if err = json.Unmarshal(unmarshal, newWalletAddress); err != nil {
					return
				}
				wallet.Addresses = append(wallet.Addresses, newWalletAddress)
				wallet.AddressesMap[string(newWalletAddress.Address.PublicKeyHash)] = newWalletAddress

				wallet.forging.Wallet.AddWallet(newWalletAddress.GetDelegatedStakePrivateKey(), newWalletAddress.GetPublicKeyHash())
				wallet.mempool.Wallet.AddWallet(newWalletAddress.GetPublicKeyHash())

			}

			wallet.updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}

func (wallet *Wallet) ReadWallet() error {

	wallet.Lock()
	defer wallet.Unlock()

	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		chainHeight, _ := binary.Uvarint(boltTx.Bucket([]byte("Chain")).Get([]byte("chainHeight")))

		accs := accounts.NewAccounts(boltTx)
		for _, adr := range wallet.Addresses {

			var acc *account.Account
			if acc, err = accs.GetAccount(adr.Address.PublicKeyHash, chainHeight); err != nil {
				return
			}

			if err = wallet.refreshWallet(acc, adr); err != nil {
				return
			}

		}

		return
	})
}
