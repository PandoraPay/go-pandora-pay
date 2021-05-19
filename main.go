package main

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/globals"
	"pandora-pay/debugging"
	"pandora-pay/gui"
	gui_interactive "pandora-pay/gui/gui-interactive"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"syscall"
)

func main() {

	var err error

	var mySettings *settings.Settings
	var myWallet *wallet.Wallet
	var myForging *forging.Forging
	var myMempool *mempool.Mempool
	var myChain *blockchain.Blockchain
	var myNetwork *network.Network

	config.StartConfig()

	if err = arguments.InitArguments(); err != nil {
		panic(err)
	}

	if globals.Arguments["--debugging"] == true {
		go debugging.Start()
	}

	if gui.GUI, err = gui_interactive.CreateGUIInteractive(); err != nil {
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

	gui.GUI.Log("Main Loop")

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	myForging.Close()
	myChain.Close()
	myWallet.Close()
	store.DBClose()

	fmt.Println("Shutting down")
}
