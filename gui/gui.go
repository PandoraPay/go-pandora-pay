package gui

import (
	"fmt"
	"pandora-pay/config"
	"pandora-pay/gui/gui_interface"
)

var GUI gui_interface.GUIInterface

func InitGUI() (err error) {

	if err = create_gui(); err != nil {
		return
	}

	GUI.Info("GO " + config.NAME)
	GUI.Info(fmt.Sprintf("OS: %s ARCH: %s %d", config.OS, config.ARCHITECTURE, config.CPU_THREADS))
	GUI.Info("VERSION " + config.VERSION_STRING)
	GUI.Info("BUILD_VERSION " + config.BUILD_VERSION)

	return
}
