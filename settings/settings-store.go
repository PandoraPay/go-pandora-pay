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

	return store.StoreSettings.DB.Update(func(boltTx *bolt.Tx) (err error) {

		writer := boltTx.Bucket([]byte("Settings"))

		writer.Put([]byte("saved"), []byte{2})

		marshal, err := json.Marshal(settings)
		if err != nil {
			return
		}

		writer.Put([]byte("settings"), marshal)
		writer.Put([]byte("saved"), []byte{1})

		return
	})

}

func (settings *Settings) loadSettings() error {
	return store.StoreSettings.DB.View(func(boltTx *bolt.Tx) (err error) {
		reader := boltTx.Bucket([]byte("Settings"))

		saved := reader.Get([]byte("saved"))
		if saved == nil {
			return errors.New("Settings doesn't exist")
		}
		if bytes.Equal(saved, []byte{1}) {
			gui.Log("Settings Loading... ")

			unmarshal := reader.Get([]byte("settings"))
			if err = json.Unmarshal(unmarshal, &settings); err != nil {
				return err
			}

			settings.updateSettings()
			gui.Log("Settings Loaded! " + settings.Name)

		} else {
			err = errors.New("Error loading wallet ?")
		}

		return
	})
}
