//go:build !wasm
// +build !wasm

package gui

import (
	"errors"
	"pandora-pay/config/arguments"
	"pandora-pay/gui/gui_interactive"
	"pandora-pay/gui/gui_non_interactive"
)

func create_gui() (err error) {

	if arguments.Arguments["--gui-type"] == "non-interactive" {
		GUI, err = gui_non_interactive.CreateGUINonInteractive()
	} else if arguments.Arguments["--gui-type"] == "interactive" {
		GUI, err = gui_interactive.CreateGUIInteractive()
	} else {
		err = errors.New("invalid --gui-type argument")
	}

	if err != nil {
		return
	}

	return
}
