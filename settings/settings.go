package settings

import (
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
	if err != nil {
		gui.Fatal("Error loading settings")
	}

	gui.Log("Settings Initialized")

}

func createEmptySettings() {
	settings = Settings{Name: helpers.RandString(10), Port: 5231}
	updateSettings()
	saveSettings()
}

func updateSettings() {
	gui.InfoUpdate("Node", settings.Name)
}
