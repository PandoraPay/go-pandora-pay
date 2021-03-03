package settings

import (
	"pandora-pay/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
)

type Settings struct {
	Name string
	Port uint16
}

var settings Settings

func SettingsInit() (err error) {

	err = loadSettings()
	if err != nil && err.Error() == "Settings doesn't exist" {
		err = createEmptySettings()
	}
	if err != nil {
		return
	}

	var changed bool
	if globals.Arguments["--node-name"] != nil {
		settings.Name = globals.Arguments["--node-name"].(string)
		changed = true
	}
	if changed {
		updateSettings()
		if err = saveSettings(); err != nil {
			return
		}
	}

	gui.Log("Settings Initialized")
	return
}

func createEmptySettings() (err error) {
	settings = Settings{Name: helpers.RandString(10), Port: 5231}
	updateSettings()

	err = saveSettings()
	if err != nil {
		return
	}

	return
}

func updateSettings() {
	gui.InfoUpdate("Node", settings.Name)
}
