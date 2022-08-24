package main

import (
	"os"
	"os/signal"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/builds/electron_helper/server"
	"pandora-pay/builds/electron_helper/server/global"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/gui"
	"syscall"
)

func main() {
	var err error

	argv := os.Args[1:]
	if err = arguments.InitArguments(argv); err != nil {
		panic(err)
	}
	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}

	if global.AddressBalanceDecryptor, err = address_balance_decryptor.NewAddressBalanceDecryptor(false); err != nil {
		return
	}

	server.CreateServer()

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

}
