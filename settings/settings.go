package settings

import (
	"pandora-pay/crypto"
	"pandora-pay/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type Settings struct {
	Name string
	Port uint16

	Checksum [4]byte

	sync.RWMutex `json:"-"`
}

func SettingsInit() (settings *Settings, err error) {

	settings = new(Settings)

	err = settings.loadSettings()
	if err != nil && err.Error() == "Settings doesn't exist" {
		err = settings.createEmptySettings()
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
	settings.Port = 5231
	settings.updateSettings()

	if err = settings.saveSettings(); err != nil {
		return
	}

	return
}

func (settings *Settings) updateSettings() {
	gui.InfoUpdate("Node", settings.Name)
}

func (settings *Settings) computeChecksum() (checksum [4]byte, err error) {

	var data []byte
	if data, err = helpers.GetJSON(settings, "Checksum"); err != nil {
		return
	}

	out := crypto.RIPEMD(data)[0:helpers.ChecksumSize]
	copy(checksum[:], out[:])

	return
}
