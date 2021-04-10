package main

import (
	"fmt"
	"github.com/docopt/docopt.go"
	"math/rand"
	"os"
	"os/signal"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/debugging"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"runtime"
	"syscall"
	"time"
)

var commands = `PANDORA PAY.

Usage:
  pandorapay [--debugging] [--version] [--testnet] [--devnet] [--debug] [--staking] [--new-devnet] [--node-name=<name>] [--tcp-server-port=<port>] [--tcp-server-address=<address>] [--tor-onion=<onion>] [--instance=<number>]
  pandorapay -h | --help

Options:
  -h --help     						Show this screen.
  --version     						Show version.
  --testnet     						Run in TESTNET mode.
  --devnet     							Run in DEVNET mode.
  --new-devnet     						Create a new devnet genesis.
  --debug     							Debug mode enabled (print log message).
  --staking     						Start staking
  --node-name=<name>   					Change node name
  --tcp-server-port=<port>				Change node tcp server port
  --tcp-server-address=<address>		Change node tcp address
  --tor-onion=<onion>					Define your tor onion address to be used.
  --instance=<number>					Number of forked instance (when you open multiple instances). It should me string number like "1","2","3","4" etc 
`

func main() {

	var err error
	var mySettings *settings.Settings
	var myWallet *wallet.Wallet
	var myForging *forging.Forging
	var myMempool *mempool.Mempool
	var myChain *blockchain.Blockchain
	var myNetwork *network.Network

	rand.Seed(time.Now().UnixNano())

	config.CPU_THREADS = runtime.GOMAXPROCS(0)
	config.ARHITECTURE = runtime.GOARCH
	config.OS = runtime.GOOS

	if globals.Arguments, err = docopt.Parse(commands, nil, false, config.VERSION, false, false); err != nil {
		panic("Error processing arguments" + err.Error())
	}

	if globals.Arguments["--debugging"] == true {
		go debugging.Start()
	}

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	if err = gui.GUIInit(); err != nil {
		panic(err)
	}
	gui.Info("GO PANDORA PAY")
	gui.Info(fmt.Sprintf("OS:%s ARCH:%s CPU:%d", config.OS, config.ARHITECTURE, config.CPU_THREADS))

	defer func() {
		err := recover()
		if err != nil {
			gui.Close()
			fmt.Print("\nERROR\n")
			fmt.Println(err)
		}
	}()

	if err = store.DBInit(); err != nil {
		panic(err)
	}

	if myMempool, err = mempool.InitMemPool(); err != nil {
		panic(err)
	}
	globals.Data["mempool"] = myMempool

	if myForging, err = forging.ForgingInit(myMempool); err != nil {
		panic(err)
	}
	globals.Data["forging"] = myForging

	if myWallet, err = wallet.WalletInit(myForging, myMempool); err != nil {
		panic(err)
	}
	globals.Data["wallet"] = myWallet

	if mySettings, err = settings.SettingsInit(); err != nil {
		panic(err)
	}
	globals.Data["settings"] = mySettings

	if myChain, err = blockchain.BlockchainInit(myForging, myWallet, myMempool); err != nil {
		panic(err)
	}
	globals.Data["chain"] = myChain

	myTransactionsBuilder := transactions_builder.TransactionsBuilderInit(myWallet, myMempool, myChain)
	globals.Data["transactionsBuilder"] = myTransactionsBuilder

	if globals.Arguments["--new-devnet"] == true {

		myTestnet := testnet.TestnetInit(myWallet, myMempool, myChain, myTransactionsBuilder)
		globals.Data["testnet"] = myTestnet

	}

	if myNetwork, err = network.CreateNetwork(mySettings, myChain, myMempool); err != nil {
		panic(err)
	}
	globals.Data["network"] = myNetwork

	gui.Log("Main Loop")

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	myForging.Close()
	myChain.Close()
	myWallet.Close()
	store.DBClose()

	fmt.Println("Shutting down")
}
