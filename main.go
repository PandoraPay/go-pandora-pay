package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"os"
	"os/signal"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/globals"
	"pandora-pay/gui"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/wallet"
	"syscall"
)

func mainloop() {

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	store.DBClose()

	fmt.Println("Shutting down")
}

var commands = `PANDORA PAY.

Usage:
  pandorapay [--version] [--testnet] [--devnet] [--debug]
  pandorapay -h | --help
  pandorapay [--node-name=<unique name>]

Options:
  -h --help     				Show this screen.
  --version     				Show version.
  --testnet     				Run in TESTNET mode.
  --devnet     					Run in DEVNET mode.
  --debug     					Debug mode enabled (print log message).
  --node-name=<unique name>   	Change node name

`

func main() {

	var err error

	gui.GUIInit()

	globals.Arguments, err = docopt.Parse(commands, nil, false, config.VERSION, false, false)
	if err != nil {
		gui.Fatal("Error processing arguments", err)
	}

	store.DBInit()

	wallet.WalletInit()
	settings.SettingsInit()

	blockchain.BlockchainInit()

	gui.Log("Main Loop")

	mainloop()
}
