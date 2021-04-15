package settings

import (
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type Settings struct {
	Name         string
	sync.RWMutex `json:"-"`
}

func SettingsInit() (settings *Settings, err error) {

	settings = &Settings{}
	if err = settings.loadSettings(); err != nil {
		if err.Error() != "Settings doesn't exist" {
			return
		}
		if err = settings.createEmptySettings(); err != nil {
			return
		}
	}

	var changed bool
	if globals.Arguments["--node-name"] != nil {
		settings.Name = globals.Arguments["--node-name"].(string)
		changed = true
	}

	if changed {
		settings.updateSettings()
		if err = settings.saveSettings(); err != nil {
			return
		}
	}

	gui.Log("Settings Initialized")
	return
}

func (settings *Settings) createEmptySettings() (err error) {
	settings.Lock()
	defer settings.Unlock()

	settings.Name = helpers.RandString(10)

	settings.updateSettings()
	if err = settings.saveSettings(); err != nil {
		return
	}
	return
}

func (settings *Settings) updateSettings() {
	gui.InfoUpdate("Node", settings.Name)
}

func (settings *Settings) computeChecksum() []byte {

	data, err := helpers.GetJSON(settings, "Checksum")
	if err != nil {
		panic(err)
	}

	return cryptography.GetChecksum(data)
}
