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

	if context.GUI, err = gui_non_interactive.CreateGUINonInteractive(); err != nil {
		panic(err)
	}
	gui.GUIInit()

	defer func() {
		err := recover()
		if err != nil {
			context.GUI.Error(err)
			context.GUI.Close()
		}
	}()

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	for i, arg := range args {
		context.GUI.Log("Argument", i, arg)
	}

}
