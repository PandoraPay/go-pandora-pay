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
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging"
	"pandora-pay/helpers/events"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"syscall"
)

var (
	mySettings *settings.Settings
	myWallet   *wallet.Wallet
	myForging  *forging.Forging
	myMempool  *mempool.Mempool
	myChain    *blockchain.Blockchain
	myNetwork  *network.Network
)

func startMain() {
	var err error

	config.StartConfig()

	argv := os.Args
	if err = arguments.InitArguments(argv); err != nil {
		panic(err)
	}

	if globals.Arguments["--debugging"] == true {
		go debugging.Start()
	}

	defer func() {
		err := recover()
		if err != nil && gui.GUI != nil {
			gui.GUI.Error(err)
			gui.GUI.Close()
			fmt.Println("Error: \n\n", err)
		}
	}()

	if err = config.InitConfig(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "config initialized")

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "GUI initialized")

	gui.GUI.Log("Arguments count", len(argv))
	for i, arg := range argv {
		gui.GUI.Log("Argument", i, arg)
	}

	if err = store.InitDB(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "database initialized")

	if myMempool, err = mempool.InitMemPool(); err != nil {
		panic(err)
	}
	globals.Data["mempool"] = myMempool
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if myForging, err = forging.ForgingInit(myMempool); err != nil {
		panic(err)
	}
	globals.Data["forging"] = myForging
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if myWallet, err = wallet.WalletInit(myForging, myMempool); err != nil {
		panic(err)
	}
	globals.Data["wallet"] = myWallet
	globals.MainEvents.BroadcastEvent("main", "wallet initialized")

	if mySettings, err = settings.SettingsInit(); err != nil {
		panic(err)
	}
	globals.Data["settings"] = mySettings
	globals.MainEvents.BroadcastEvent("main", "settings initialized")

	if myChain, err = blockchain.BlockchainInit(myForging, myWallet, myMempool); err != nil {
		panic(err)
	}
	globals.Data["chain"] = myChain
	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	myTransactionsBuilder := transactions_builder.TransactionsBuilderInit(myWallet, myMempool, myChain)
	globals.Data["transactionsBuilder"] = myTransactionsBuilder
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	if globals.Arguments["--new-devnet"] == true {

		myTestnet := testnet.TestnetInit(myWallet, myMempool, myChain, myTransactionsBuilder)
		globals.Data["testnet"] = myTestnet

	}

	if myNetwork, err = network.CreateNetwork(mySettings, myChain, myMempool); err != nil {
		panic(err)
	}
	globals.Data["network"] = myNetwork
	globals.MainEvents.BroadcastEvent("main", "network initialized")

	gui.GUI.Log("Main Loop")
	globals.MainEvents.BroadcastEvent("main", "initialized")

}

func main() {

	globals.MainEvents = events.NewEvents()

	additionalMain()

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	myForging.Close()
	myChain.Close()
	myWallet.Close()
	store.DBClose()

	fmt.Println("Shutting down")
}
