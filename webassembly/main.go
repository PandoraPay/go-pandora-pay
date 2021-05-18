package main

import (
	"fmt"
	"pandora-pay/config"
	"pandora-pay/gui"
	gui_non_interactive "pandora-pay/gui/gui-non-interactive"
)

func main() {
	var err error

	config.StartConfig()

	if gui.GUI, err = gui_non_interactive.CreateGUINonInteractive(); err != nil {
		panic(err)
	}
	gui.GUIInit()

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	defer func() {
		err := recover()
		if err != nil {
			gui.GUI.Close()
			fmt.Print("\nERROR\n")
			fmt.Println(err)
		}
	}()
}
