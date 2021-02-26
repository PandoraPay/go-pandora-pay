package wallet

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/crypto"
	"pandora-pay/forging"
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

var wSaved = WalletSaved{}

func saveWallet() error {
	return store.StoreWallet.DB.Update(func(tx *bolt.Tx) (err error) {

		var marshal, checksum []byte

		writer := tx.Bucket([]byte("Wallet"))

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return gui.Error("Error deleting saved status", err)
		}

		if marshal, err = json.Marshal(wSaved); err != nil {
			return gui.Error("Error marshaling wallet saved", err)
		}
		checksum = append(checksum, marshal...)
		if err = writer.Put([]byte("wallet-saved"), marshal); err != nil {
			return gui.Error("Error storing saved status", err)
		}

		if marshal, err = json.Marshal(W); err != nil {
			return gui.Error("Error marshaling wallet", err)
		}

		checksum = append(checksum, marshal...)
		if err = writer.Put([]byte("wallet"), marshal); err != nil {
			return gui.Error("Error storing saved status", err)
		}

		for i := 0; i < W.Count; i++ {
			if marshal, err = json.Marshal(W.Addresses[i]); err != nil {
				return gui.Error("Error marshaling address "+strconv.Itoa(i), err)
			}
			checksum = append(checksum, marshal...)
			err = writer.Put([]byte("wallet-address-"+strconv.Itoa(i)), marshal)
		}

		if err = writer.Delete([]byte("wallet-address-" + strconv.Itoa(W.Count))); err != nil {
			return gui.Error("Error deleting next address", err)
		}

		checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
		if err = writer.Put([]byte("wallet-check-sum"), checksum); err != nil {
			return gui.Error("Error storing checksum", err)
		}

		if err = writer.Put([]byte("saved"), []byte{1}); err != nil {
			return gui.Error("Error storing final wallet saved", err)
		}

		return nil
	})
}

func loadWallet() error {

	return store.StoreWallet.DB.View(func(tx *bolt.Tx) (err error) {

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

			if err = json.Unmarshal(unmarshal, &wSaved); err != nil {
				return gui.Error("Error unmarshaling wallet saved", err)
			}

			unmarshal = reader.Get([]byte("wallet"))
			checksum = append(checksum, unmarshal...)
			if err = json.Unmarshal(unmarshal, &newWallet); err != nil {
				return gui.Error("Error unmarshaling wallet", err)
			}

			for i := 0; i < newWallet.Count; i++ {
				unmarshal = reader.Get([]byte("wallet-address-" + strconv.Itoa(i)))
				checksum = append(checksum, unmarshal...)

				newWalletAddress := WalletAddress{}
				if err = json.Unmarshal(unmarshal, &newWalletAddress); err != nil {
					return gui.Error("Error unmarshaling address "+strconv.Itoa(i), err)
				}
				newWallet.Addresses = append(newWallet.Addresses, &newWalletAddress)
				go forging.ForgingW.AddWallet(newWalletAddress.PublicKey, newWalletAddress.PrivateKey.Key, newWalletAddress.PublicKeyHash)
			}

			checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
			walletChecksum := reader.Get([]byte("wallet-check-sum"))
			if !bytes.Equal(checksum, walletChecksum) {
				return gui.Error("Wallet Checksum is not matching", errors.New("Wallet checksum mismatch !"))
			}

			W = newWallet
			updateWallet()
			gui.Log("Wallet Loaded! " + strconv.Itoa(W.Count))

		} else {
			gui.Fatal("Error loading wallet ?")
		}
		return nil
	})
}
