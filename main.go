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
	"pandora-pay/testnet"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"runtime"
	"syscall"
)

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

	config.InitConfig()
	store.DBInit()

	myMempool := mempool.InitMemPool()
	globals.Data["mempool"] = myMempool

	myForging := forging.ForgingInit(myMempool)
	globals.Data["forging"] = myForging

	myWallet := wallet.WalletInit(myForging)
	globals.Data["wallet"] = myWallet

	mySettings := settings.SettingsInit()
	globals.Data["settings"] = mySettings

	myChain := blockchain.BlockchainInit(myForging, myMempool)
	globals.Data["chain"] = myChain

	myTransactionsBuilder := transactions_builder.TransactionsBuilderInit(myWallet, myChain)
	globals.Data["transactionsBuilder"] = myTransactionsBuilder

	if globals.Arguments["--new-genesis"] == true {

		myTestnet := testnet.TestnetInit(myWallet, myMempool, myChain, myTransactionsBuilder)
		globals.Data["testnet"] = myTestnet

	}

	gui.Log("Main Loop")

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	myChain.Close()
	myForging.Close()
	store.DBClose()

	fmt.Println("Shutting down")
}
