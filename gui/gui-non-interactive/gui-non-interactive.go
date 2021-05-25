package gui_non_interactive

import (
	gui_interface "pandora-pay/gui/gui-interface"
	gui_logger "pandora-pay/gui/gui-logger"
	"runtime"
)

type GUINonInteractive struct {
	gui_interface.GUIInterface
	logger       *gui_logger.GUILogger
	colorError   string
	colorWarning string
	colorInfo    string
	colorLog     string
	colorFatal   string
}

func (g *GUINonInteractive) Close() {
}

func CreateGUINonInteractive() (g *GUINonInteractive, err error) {

	g = &GUINonInteractive{}

	switch runtime.GOARCH {
	default:
		g.colorError = "\x1b[31m"
		g.colorWarning = "\x1b[32m"
		g.colorInfo = "\x1b[34m"
		g.colorLog = "\x1b[37m"
		g.colorFatal = "\x1b[31m\x1b[43m"
	}

	return
}

func (g *GUINonInteractive) InfoUpdate(key string, text string) {
}

func (g *GUINonInteractive) Info2Update(key string, text string) {
}

func (g *GUINonInteractive) OutputWrite(any interface{}) {
}

func (g *GUINonInteractive) CommandDefineCallback(Text string, callback func(string) error) {
}
