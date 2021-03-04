package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/gui"
	"pandora-pay/store"
)

func (settings *Settings) saveSettings() error {

	return store.StoreSettings.DB.Update(func(tx *bolt.Tx) (err error) {

		var marshal []byte
		writer := tx.Bucket([]byte("Settings"))

		if settings.Checksum, err = settings.computeChecksum(); err != nil {
			return
		}

		if err = writer.Put([]byte("saved"), []byte{2}); err != nil {
			return
		}
		if marshal, err = json.Marshal(settings); err != nil {
			return
		}
		if err = writer.Put([]byte("settings"), marshal); err != nil {
			return
		}
		if err = writer.Put([]byte("saved"), []byte{1}); err != nil {
			return
		}

		return nil
	})

}

func (settings *Settings) loadSettings() error {

	return store.StoreSettings.DB.View(func(tx *bolt.Tx) (err error) {

		reader := tx.Bucket([]byte("Settings"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Settings doesn't exist")
		}
		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Settings Loading... ")

			unmarshal := reader.Get([]byte("settings"))
			if err = json.Unmarshal(unmarshal, &settings); err != nil {
				return errors.New("Error unmarshaling settings saved")
			}

			var checksum [4]byte
			if checksum, err = settings.computeChecksum(); err != nil {
				return
			}
			if !bytes.Equal(checksum[:], settings.Checksum[:]) {
				return errors.New("Settings checksum mismatch !")
			}

			settings.updateSettings()
			gui.Log("Settings Loaded! " + settings.Name)

		} else {
			return errors.New("Error loading wallet ?")
		}

		return nil
	})

}
