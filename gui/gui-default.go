// +build !wasm

package gui

import (
	gui_interactive "pandora-pay/gui/gui-interactive"
)

func create_gui() (err error) {
	if GUI, err = gui_interactive.CreateGUIInteractive(); err != nil {
		return
	}
	return
}
