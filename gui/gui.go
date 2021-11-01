package gui

import (
	"pandora-pay/config"
	"pandora-pay/gui/gui_interface"
	"strconv"
)

var GUI gui_interface.GUIInterface

//test
func InitGUI() (err error) {

	if err = create_gui(); err != nil {
		return
	}

	GUI.Info("GO " + config.NAME)
	GUI.Info("OS: " + config.OS + "ARCH: " + config.ARCHITECTURE + "CPU: " + strconv.Itoa(config.CPU_THREADS))
	GUI.Info("VERSION " + config.VERSION)
	GUI.Info("BUILD_VERSION " + config.BUILD_VERSION)

	return
}
