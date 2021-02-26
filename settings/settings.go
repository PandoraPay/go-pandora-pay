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

func SettingsInit() {

	err := loadSettings()
	if err != nil && err.Error() == "Settings doesn't exist" {
		err = createEmptySettings()
	}
	if err != nil {
		gui.Fatal("Error loading settings", err)
	}

	var changed bool
	if globals.Arguments["--node-name"] != nil {
		settings.Name = globals.Arguments["--node-name"].(string)
		changed = true
	}
	if changed {
		updateSettings()
		err = saveSettings()
		if err != nil {
			gui.Fatal("Error saving new", err)
		}
	}

	gui.Log("Settings Initialized")

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
