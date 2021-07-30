package settings

import (
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type Settings struct {
	Name         string `json:"name"`
	sync.RWMutex `json:"-"`
}

func SettingsInit() (*Settings, error) {

	settings := &Settings{}
	if err := settings.loadSettings(); err != nil {
		if err.Error() != "Settings doesn't exist" {
			return nil, err
		}
		if err = settings.createEmptySettings(); err != nil {
			return nil, err
		}
	}

	var changed bool
	if globals.Arguments["--node-name"] != nil {
		settings.Name = globals.Arguments["--node-name"].(string)
		changed = true
	}

	if changed {
		settings.updateSettings()
		if err := settings.saveSettings(); err != nil {
			return nil, err
		}
	}

	gui.GUI.Log("Settings Initialized")
	return settings, nil
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
	gui.GUI.InfoUpdate("Node", settings.Name)
}
