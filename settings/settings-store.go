package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"pandora-pay/store"
)

func saveSettings() error {

	return store.StoreSettings.DB.Update(func(tx *bolt.Tx) error {

		var marshal, checksum []byte

		writer := tx.Bucket([]byte("Settings"))

		err := writer.Put([]byte("saved"), []byte{2})
		if err != nil {
			return gui.Error("Error deleting saved status", err)
		}

		marshal, err = json.Marshal(settings)
		if err != nil {
			return gui.Error("Error marshaling wallet saved", err)
		}
		checksum = append(checksum, marshal...)
		err = writer.Put([]byte("settings"), marshal)
		if err != nil {
			return gui.Error("Error storing saved status", err)
		}

		checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
		err = writer.Put([]byte("settings-check-sum"), checksum)
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

func loadSettings() error {

	return store.StoreSettings.DB.View(func(tx *bolt.Tx) error {

		reader := tx.Bucket([]byte("Settings"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return createEmptySettings()
		}
		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Settings Loading... ")

			var unmarshal, checksum []byte
			newSettings := Settings{}

			unmarshal = reader.Get([]byte("settings"))
			checksum = append(checksum, unmarshal...)
			err := json.Unmarshal(unmarshal, &newSettings)
			if err != nil {
				return gui.Error("Error unmarshaling wallet saved", err)
			}

			checksum = crypto.RIPEMD(checksum)[0:crypto.ChecksumSize]
			walletChecksum := reader.Get([]byte("settings-check-sum"))
			if !bytes.Equal(checksum, walletChecksum) {
				return gui.Error("Settings Checksum is not matching", errors.New("Settings checksum mismatch !"))
			}

			settings = newSettings
			updateSettings()

			gui.Log("Settings Loaded! " + settings.Name)

		} else {
			gui.Fatal("Error loading wallet ?")
		}

		return nil
	})

}
