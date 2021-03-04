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
	"pandora-pay/mempool"
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

	blockchain.BlockchainClose()
	store.DBClose()

	fmt.Println("Shutting down")
}

var commands = `PANDORA PAY.

Usage:
  pandorapay [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-genesis] [--node-name=<name>]
  pandorapay -h | --help

Options:
  -h --help     				Show this screen.
  --version     				Show version.
  --testnet     				Run in TESTNET mode.
  --devnet     					Run in DEVNET mode.
  --new-genesis     			Create a new genesis.
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

	if globals.Arguments, err = docopt.Parse(commands, nil, false, config.VERSION, false, false); err != nil {
		gui.Fatal("Error processing arguments", err)
	}

	if err = config.InitConfig(); err != nil {
		gui.Fatal("Error initializing Config", err)
	}

	if err = store.DBInit(); err != nil {
		gui.Fatal("Error initializing Database", err)
	}

	forging, err := forging.ForgingInit()
	if err != nil {
		gui.Fatal("Error initializing Forging", err)
	}
	globals.Data["forging"] = forging

	wallet, err := wallet.WalletInit(forging)
	if err != nil {
		gui.Fatal("Error initializing Wallet", err)
	}
	globals.Data["wallet"] = wallet

	settings, err := settings.SettingsInit()
	if err != nil {
		gui.Fatal("Error initializing Settings", err)
	}
	globals.Data["settings"] = settings

	chain, err := blockchain.BlockchainInit(forging)
	if err != nil {
		gui.Fatal("Error Initializing Blockchain", err)
	}
	globals.Data["chain"] = chain

	mempool, err := mempool.InitMemPool()
	if err != nil {
		gui.Fatal("Error initializing Mempool", err)
	}
	globals.Data["mempool"] = mempool

	go func() {

		for {
			_ = <-chain.UpdateChannel
		}

	}()

	gui.Log("Main Loop")

	mainloop()
}
