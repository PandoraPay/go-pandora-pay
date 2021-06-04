package start

import (
	"fmt"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
)

func startMain() {

	if globals.MainStarted {
		return
	}
	globals.MainStarted = true

	var err error

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

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "GUI initialized")

	if err = store.InitDB(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "database initialized")

	if app.Mempool, err = mempool.InitMemPool(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if app.Chain, err = blockchain.BlockchainInit(app.Mempool); err != nil {
		panic(err)
	}

	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	if app.Forging, err = forging.ForgingInit(app.Mempool, app.Chain.NextBlockCreatedCn, app.Chain.UpdateAccounts, app.Chain.ForgingSolutionCn); err != nil {
		panic(err)
	}

	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if app.Wallet, err = wallet.WalletInit(app.Forging, app.Mempool, app.Chain.UpdateAccounts); err != nil {
		panic(err)
	}

	globals.MainEvents.BroadcastEvent("main", "wallet initialized")

	if err = app.Wallet.InitializeWallet(); err != nil {
		return
	}
	app.Chain.InitForging()
	app.Forging.StartForging()

	if app.Settings, err = settings.SettingsInit(); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "settings initialized")

	app.TransactionsBuilder = transactions_builder.TransactionsBuilderInit(app.Wallet, app.Mempool, app.Chain)
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	if globals.Arguments["--new-devnet"] == true {

		myTestnet := testnet.TestnetInit(app.Wallet, app.Mempool, app.Chain, app.TransactionsBuilder)
		globals.Data["testnet"] = myTestnet

	}

	if app.Network, err = network.CreateNetwork(app.Settings, app.Chain, app.Mempool); err != nil {
		panic(err)
	}
	globals.MainEvents.BroadcastEvent("main", "network initialized")

	gui.GUI.Log("Main Loop")
	globals.MainEvents.BroadcastEvent("main", "initialized")

}

func CloseMain() {
	gui.GUI.Close()
	app.Forging.Close()
	app.Chain.Close()
	app.Wallet.Close()
	store.DBClose()
}
