package gui

import (
	"fmt"
	"pandora-pay/config"
	gui_interface "pandora-pay/gui/gui-interface"
)

var GUI gui_interface.GUIInterface

//test
func GUIInit() {

	GUI.Info("GO PANDORA PAY")
	GUI.Info(fmt.Sprintf("OS:%s ARCH:%s CPU:%d", config.OS, config.ARCHITECTURE, config.CPU_THREADS))

	return
}
