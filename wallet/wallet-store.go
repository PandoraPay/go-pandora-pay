package wallet

import (
	"bytes"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"strconv"
)

func (wallet *Wallet) saveWallet(start, end, deleteIndex int) {

	if err := store.StoreWallet.DB.Update(func(tx *bolt.Tx) error {

		writer := tx.Bucket([]byte("Wallet"))
		var marshal []byte

		wallet.Checksum = wallet.computeChecksum()

		writer.Put([]byte("saved"), []byte{2})

		marshal, err := helpers.GetJSON(wallet, "Addresses")
		if err != nil {
			panic(err)
		}

		writer.Put([]byte("wallet"), marshal)

		for i := start; i < end; i++ {
			if marshal, err = json.Marshal(wallet.Addresses[i]); err != nil {
				panic(err)
			}
			writer.Put([]byte("wallet-address-"+strconv.Itoa(i)), marshal)
		}

		if deleteIndex != -1 {
			writer.Delete([]byte("wallet-address-" + strconv.Itoa(deleteIndex)))
		}

		writer.Put([]byte("saved"), []byte{1})

		return nil
	}); err != nil {
		panic(err)
	}
}

func (wallet *Wallet) loadWallet() {

	if err := store.StoreWallet.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Wallet"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			panic("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {

			gui.Log("Wallet Loading... ")

			var unmarshal []byte

			unmarshal = reader.Get([]byte("wallet"))

			if err := json.Unmarshal(unmarshal, &wallet); err != nil {
				panic("Error unmarshaling wallet")
			}

			for i := 0; i < wallet.Count; i++ {
				unmarshal = reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))

				newWalletAddress := WalletAddress{}
				if err := json.Unmarshal(unmarshal, &newWalletAddress); err != nil {
					panic("Error unmarshaling address " + strconv.Itoa(i))
				}
				wallet.Addresses = append(wallet.Addresses, &newWalletAddress)
				go wallet.forging.Wallet.AddWallet(newWalletAddress.PublicKey, newWalletAddress.PrivateKey.Key, newWalletAddress.PublicKeyHash)
			}

			checksum := wallet.computeChecksum()
			if !bytes.Equal(checksum, wallet.Checksum) {
				panic("Wallet checksum mismatch !")
			}

			wallet.updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			panic("Error loading wallet ?")
		}
		return nil
	}); err != nil {
		panic(err)
	}
}
