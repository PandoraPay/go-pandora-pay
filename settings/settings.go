package settings

import (
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type Settings struct {
	Name string
	Port uint16

	Checksum cryptography.Checksum

	sync.RWMutex `json:"-"`
}

func SettingsInit() (settings *Settings) {

	defer func() {
		if err := recover(); err != nil {
			if helpers.ConvertRecoverError(err).Error() == "Settings doesn't exist" {
				settings.createEmptySettings()
			} else {
				panic(err)
			}
		}
	}()

	settings = &Settings{}
	settings.loadSettings()

	var changed bool
	if globals.Arguments["--node-name"] != nil {
		settings.Name = globals.Arguments["--node-name"].(string)
		changed = true
	}
	if changed {
		settings.updateSettings()
		settings.saveSettings()
	}

	gui.Log("Settings Initialized")
	return
}

func (settings *Settings) createEmptySettings() {

	settings.Lock()
	defer settings.Unlock()

	settings.Name = helpers.RandString(10)
	settings.Port = 5231
	settings.updateSettings()

	settings.saveSettings()

	return
}

func (settings *Settings) updateSettings() {
	gui.InfoUpdate("Node", settings.Name)
}

func (settings *Settings) computeChecksum() (checksum cryptography.Checksum) {

	data, err := helpers.GetJSON(settings, "Checksum")
	if err != nil {
		panic(err)
	}

	out := cryptography.GetChecksum(data)
	copy(checksum[:], out[:])
	return

}
