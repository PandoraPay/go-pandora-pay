//go:build wasm
// +build wasm

package gui

import gui_non_interactive "pandora-pay/gui/gui_non_interactive"

func create_gui() (err error) {
	if GUI, err = gui_non_interactive.CreateGUINonInteractive(); err != nil {
		return
	}
	return
}
