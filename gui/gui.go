package gui

import (
	"fmt"
	"pandora-pay/config"
	"pandora-pay/context"
)

//test
func GUIInit() {
	context.GUI.Info("GO " + config.NAME)
	context.GUI.Info(fmt.Sprintf("OS:%s ARCH:%s CPU:%d", config.OS, config.ARCHITECTURE, config.CPU_THREADS))
	context.GUI.Info("VERSION " + config.VERSION)
}
