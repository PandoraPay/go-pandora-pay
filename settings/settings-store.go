package settings

import (
	"bytes"
	"encoding/json"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/gui"
	"pandora-pay/store"
)

func (settings *Settings) saveSettings() {

	if err := store.StoreSettings.DB.Update(func(tx *bolt.Tx) error {

		writer := tx.Bucket([]byte("Settings"))

		settings.Checksum = settings.computeChecksum()

		writer.Put([]byte("saved"), []byte{2})

		marshal, err := json.Marshal(settings)
		if err != nil {
			panic(err)
		}

		writer.Put([]byte("settings"), marshal)
		writer.Put([]byte("saved"), []byte{1})

		return nil
	}); err != nil {
		panic(err)
	}

}

func (settings *Settings) loadSettings() {

	if err := store.StoreSettings.DB.View(func(tx *bolt.Tx) (err error) {

		reader := tx.Bucket([]byte("Settings"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			panic("Settings doesn't exist")
		}
		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Settings Loading... ")

			unmarshal := reader.Get([]byte("settings"))
			if err = json.Unmarshal(unmarshal, &settings); err != nil {
				panic("Error unmarshaling settings saved")
			}

			checksum := settings.computeChecksum()

			if !bytes.Equal(checksum[:], settings.Checksum[:]) {
				panic("Settings checksum mismatch !")
			}

			settings.updateSettings()
			gui.Log("Settings Loaded! " + settings.Name)

		} else {
			panic("Error loading wallet ?")
		}

		return nil
	}); err != nil {
		panic(err)
	}

}
