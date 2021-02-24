package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"os"
	"os/signal"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/forging"
	"pandora-pay/globals"
	"pandora-pay/gui"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/wallet"
	"runtime"
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
  pandorapay [--version] [--testnet] [--devnet] [--debug] [--staking] [--node-name=<name>]
  pandorapay -h | --help

Options:
  -h --help     				Show this screen.
  --version     				Show version.
  --testnet     				Run in TESTNET mode.
  --devnet     					Run in DEVNET mode.
  --debug     					Debug mode enabled (print log message).
  --staking     				Start staking
  --node-name=<name>   			Change node name

`

func main() {

	var err error

	gui.GUIInit()
	gui.Info("GO PANDORA PAY")

	config.CPU_THREADS = runtime.GOMAXPROCS(0)
	config.ARHITECTURE = runtime.GOARCH
	config.OS = runtime.GOOS

	gui.Info(fmt.Sprintf("OS:%s ARCH:%s CPU:%d", config.OS, config.ARHITECTURE, config.CPU_THREADS))

	globals.Arguments, err = docopt.Parse(commands, nil, false, config.VERSION, false, false)
	if err != nil {
		gui.Fatal("Error processing arguments", err)
	}

	config.InitConfig()

	store.DBInit()

	wallet.WalletInit()
	settings.SettingsInit()

	blockchain.BlockchainInit()

	forging.ForgingInit()

	gui.Log("Main Loop")

	mainloop()
}
