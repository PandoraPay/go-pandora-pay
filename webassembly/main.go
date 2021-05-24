package main

import (
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/store"
	"strings"
	"syscall/js"
)

func main() {

	var err error
	var myMempool *mempool.Mempool

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

	defer func() {
		err := recover()
		if err != nil && gui.GUI != nil {
			gui.GUI.Error(err)
			gui.GUI.Close()
		}
	}()

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}

	for i, arg := range args {
		gui.GUI.Log("Argument", i, arg)
	}

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	if err = store.InitDB(); err != nil {
		panic(err)
	}

	if myMempool, err = mempool.InitMemPool(); err != nil {
		panic(err)
	}
	globals.Data["mempool"] = myMempool

}
