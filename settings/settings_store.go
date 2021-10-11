package settings

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/gui"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
)

func (settings *Settings) saveSettings() error {

	return store.StoreSettings.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

		writer.Put("saved", []byte{2})

		marshal, err := json.Marshal(settings)
		if err != nil {
			return
		}

		writer.Put("settings", marshal)
		writer.Put("saved", []byte{1})

		return
	})

}

func (settings *Settings) loadSettings() error {
	return store.StoreSettings.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		saved := reader.Get("saved")
		if saved == nil {
			return errors.New("Settings doesn't exist")
		}
		if bytes.Equal(saved, []byte{1}) {
			gui.GUI.Log("Settings Loading... ")

			unmarshal := reader.Get("settings")
			if err = json.Unmarshal(unmarshal, &settings); err != nil {
				return err
			}

			settings.updateSettings()
			gui.GUI.Log("Settings Loaded! " + settings.Name)

		} else {
			err = errors.New("Error loading wallet ?")
		}

		return
	})
}
