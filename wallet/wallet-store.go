package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (wallet *Wallet) saveWallet(start, end, deleteIndex int) error {
	return store.StoreWallet.DB.Update(func(boltTx *bolt.Tx) (err error) {

		writer := boltTx.Bucket([]byte("Wallet"))

		wallet.Checksum = wallet.computeChecksum()

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return
		}

		marshal, err := helpers.GetJSON(wallet, "Addresses")
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

				newWalletAddress := WalletAddress{}
				if err = json.Unmarshal(unmarshal, &newWalletAddress); err != nil {
					return
				}
				wallet.Addresses = append(wallet.Addresses, &newWalletAddress)

				go wallet.forging.Wallet.AddWallet(newWalletAddress.PrivateKey.Key, newWalletAddress.PublicKeyHash)
				go wallet.mempool.Wallet.AddWallet(newWalletAddress.PublicKeyHash)
			}

			checksum := wallet.computeChecksum()
			if !bytes.Equal(checksum, wallet.Checksum) {
				return errors.New("Wallet checksum mismatch !")
			}

			wallet.updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			return errors.New("Error loading wallet ?")
		}
		return
	})
}
