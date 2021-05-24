package main

import (
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/context"
	"pandora-pay/gui"
	gui_non_interactive "pandora-pay/gui/gui-non-interactive"
	"strings"
	"syscall/js"
)

func main() {
	var err error

	config.StartConfig()

	args := []string{}
	jsConfig := js.Global().Get("PandoraPayConfig")
	if jsConfig.Truthy() {
		if jsConfig.Type() != js.TypeString {
			panic("PandoraPayConfig must be a string")
		}
		args = strings.Split(jsConfig.String(), " ")
	}

	if err = arguments.InitArguments(args); err != nil {
		panic(err)
	}

	if gui.GUI, err = gui_non_interactive.CreateGUINonInteractive(); err != nil {
		panic(err)
	}
	gui.GUIInit()

	defer func() {
		err := recover()
		if err != nil {
			gui.GUI.Error(err)
			gui.GUI.Close()
		}
	}()

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	for i, arg := range args {
		gui.GUI.Log("Argument", i, arg)
	}

}
