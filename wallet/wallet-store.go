package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/store"
	"strconv"
)

type EncryptedVersion int

const (
	PlainText EncryptedVersion = 0
	Encrypted EncryptedVersion = 1
)

func (e EncryptedVersion) String() string {
	switch e {
	case PlainText:
		return "PlainText"
	case Encrypted:
		return "Encrypted"
	default:
		return "Unknown EncryptedVersion"
	}
}

type WalletSaved struct {
	Saved     bool
	Encrypted EncryptedVersion
	Checksum  [4]byte
}

var walletSaved = WalletSaved{}

func saveWallet() error {
	return store.StoreWallet.DB.Update(func(tx *bolt.Tx) error {

		var marshal, checksum []byte

		writer := tx.Bucket([]byte("Wallet"))

		err := writer.Put([]byte("saved"), []byte{2})
		if err != nil {
			return gui.Error("Error deleting saved status", err)
		}

		marshal, err = json.Marshal(walletSaved)
		if err != nil {
			return gui.Error("Error marshaling wallet saved", err)
		}
		checksum = append(checksum, marshal...)
		err = writer.Put([]byte("wallet-saved"), marshal)
		if err != nil {
			return gui.Error("Error storing saved status", err)
		}

		marshal, err = json.Marshal(wallet)
		if err != nil {
			return gui.Error("Error marshaling wallet", err)
		}
		checksum = append(checksum, marshal...)
		err = writer.Put([]byte("wallet"), marshal)
		if err != nil {
			return gui.Error("Error storing saved status", err)
		}

		for i := 0; i < wallet.Count; i++ {
			marshal, err = json.Marshal(wallet.Addresses[i])
			checksum = append(checksum, marshal...)
			if err != nil {
				return gui.Error("Error marshaling address "+strconv.Itoa(i), err)
			}
			err = writer.Put([]byte("wallet-address-"+strconv.Itoa(i)), marshal)
		}

		err = writer.Delete([]byte("wallet-address-" + strconv.Itoa(wallet.Count)))
		if err != nil {
			return gui.Error("Error deleting next address", err)
		}

		checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
		err = writer.Put([]byte("wallet-check-sum"), checksum)
		if err != nil {
			return gui.Error("Error storing checksum", err)
		}

		err = writer.Put([]byte("saved"), []byte{1})
		if err != nil {
			return gui.Error("Error storing final wallet saved", err)
		}

		return nil
	})
}

func loadWallet() error {

	return store.StoreWallet.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Wallet"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Wallet doesn't exist")
		}

		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Wallet Loading... ")

			var unmarshal, checksum []byte
			newWallet := Wallet{}

			unmarshal = reader.Get([]byte("wallet-saved"))
			checksum = append(checksum, unmarshal...)
			err := json.Unmarshal(unmarshal, &walletSaved)
			if err != nil {
				return gui.Error("Error unmarshaling wallet saved", err)
			}

			unmarshal = reader.Get([]byte("wallet"))
			checksum = append(checksum, unmarshal...)
			err = json.Unmarshal(unmarshal, &newWallet)
			if err != nil {
				return gui.Error("Error unmarshaling wallet", err)
			}

			for i := 0; i < newWallet.Count; i++ {
				unmarshal = reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))
				checksum = append(checksum, unmarshal...)

				newWalletAddress := WalletAddress{}
				err := json.Unmarshal(unmarshal, &newWalletAddress)
				if err != nil {
					return gui.Error("Error unmarshaling address "+strconv.Itoa(i), err)
				}
				newWallet.Addresses = append(newWallet.Addresses, &newWalletAddress)
			}

			checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
			walletChecksum := reader.Get([]byte("wallet-check-sum"))
			if !bytes.Equal(checksum, walletChecksum) {
				return gui.Error("Wallet Checksum is not matching", errors.New("Wallet checksum mismatch !"))
			}

			wallet = newWallet
			updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(wallet.Count))

		} else {
			gui.Fatal("Error loading wallet ?")
		}
		return nil
	})
}
