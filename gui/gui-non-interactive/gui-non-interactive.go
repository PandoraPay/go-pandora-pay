package gui_non_interactive

import (
	gui_interface "pandora-pay/gui/gui-interface"
	gui_logger "pandora-pay/gui/gui-logger"
)

type GUINonInteractive struct {
	gui_interface.GUIInterface
	logger *gui_logger.GUILogger
}

func (g *GUINonInteractive) Close() {

}

func CreateGUINonInteractive() (g *GUINonInteractive, err error) {

	g = &GUINonInteractive{}

	return
}
