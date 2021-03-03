package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/blockchain/forging"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

func (wallet *Wallet) saveWallet() error {
	return store.StoreWallet.DB.Update(func(tx *bolt.Tx) (err error) {

		if wallet.Checksum, err = wallet.computeChecksum(); err != nil {
			return
		}

		writer := tx.Bucket([]byte("Wallet"))
		var marshal []byte

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return gui.Error("Error deleting saved status", err)
		}

		if marshal, err = json.Marshal(wallet); err != nil {
			return gui.Error("Error marshaling wallet", err)
		}

		if err = writer.Put([]byte("wallet"), marshal); err != nil {
			return gui.Error("Error storing saved status", err)
		}

		for i := 0; i < wallet.Count; i++ {
			if marshal, err = json.Marshal(wallet.Addresses[i]); err != nil {
				return gui.Error("Error marshaling address "+strconv.Itoa(i), err)
			}
			err = writer.Put([]byte("wallet-address-"+strconv.Itoa(i)), marshal)
		}

		if err = writer.Delete([]byte("wallet-address-" + strconv.Itoa(wallet.Count))); err != nil {
			return gui.Error("Error deleting next address", err)
		}

		if err = writer.Put([]byte("wallet-checksum"), wallet.Checksum[:]); err != nil {
			return gui.Error("Error storing checksum", err)
		}

		if err = writer.Put([]byte("saved"), []byte{1}); err != nil {
			return gui.Error("Error storing final wallet saved", err)
		}

		return nil
	})
}

func (wallet *Wallet) loadWallet() error {

	return store.StoreWallet.DB.View(func(tx *bolt.Tx) (err error) {

		reader := tx.Bucket([]byte("Wallet"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Wallet Loading... ")

			var unmarshal []byte

			unmarshal = reader.Get([]byte("wallet"))

			if err = json.Unmarshal(unmarshal, &wallet); err != nil {
				return gui.Error("Error unmarshaling wallet", err)
			}

			for i := 0; i < wallet.Count; i++ {
				unmarshal = reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))

				newWalletAddress := WalletAddress{}
				if err = json.Unmarshal(unmarshal, &newWalletAddress); err != nil {
					return gui.Error("Error unmarshaling address "+strconv.Itoa(i), err)
				}
				wallet.Addresses = append(wallet.Addresses, &newWalletAddress)
				go forging.ForgingW.AddWallet(newWalletAddress.PublicKey, newWalletAddress.PrivateKey.Key, newWalletAddress.PublicKeyHash)
			}

			unmarshal = reader.Get([]byte("wallet-checksum"))
			copy(wallet.Checksum[:], unmarshal[:])

			var checksum [4]byte
			if checksum, err = wallet.computeChecksum(); err != nil {
				return
			}
			if !bytes.Equal(checksum[:], wallet.Checksum[:]) {
				return gui.Error("Wallet Checksum is not matching", errors.New("Wallet checksum mismatch !"))
			}

			wallet.updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			gui.Fatal("Error loading wallet ?")
		}
		return nil
	})
}
