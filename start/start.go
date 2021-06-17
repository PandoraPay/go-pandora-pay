package start

import (
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/recovery"
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
		recovery.SafeGo(debugging.Start)
	}

	if err = gui.InitGUI(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "GUI initialized")

	if err = store.InitDB(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "database initialized")

	if app.Mempool, err = mempool.CreateMemPool(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if app.Forging, err = forging.CreateForging(app.Mempool); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if app.Chain, err = blockchain.CreateBlockchain(app.Mempool); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	app.Forging.InitializeForging(app.Chain.NextBlockCreatedCn, app.Chain.UpdateAccounts, app.Chain.ForgingSolutionCn)

	if app.Wallet, err = wallet.CreateWallet(app.Forging, app.Mempool); err != nil {
		return
	}
	if err = app.Wallet.ProcessWalletArguments(); err != nil {
		return
	}

	globals.MainEvents.BroadcastEvent("main", "wallet initialized")
	app.Wallet.InitializeWallet(app.Chain.UpdateAccounts)

	if err = genesis.GenesisInit(app.Wallet); err != nil {
		return
	}
	if err = app.Chain.InitializeChain(); err != nil {
		return
	}
	if err = app.Wallet.StartWallet(); err != nil {
		return
	}

	app.Forging.StartForging()

	app.Chain.InitForging()

	if app.Settings, err = settings.SettingsInit(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "settings initialized")

	app.TransactionsBuilder = transactions_builder.TransactionsBuilderInit(app.Wallet, app.Mempool, app.Chain)
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	if globals.Arguments["--new-devnet"] == true {
		myTestnet := testnet.TestnetInit(app.Wallet, app.Mempool, app.Chain, app.TransactionsBuilder)
		globals.Data["testnet"] = myTestnet
	}

	if app.Network, err = network.CreateNetwork(app.Settings, app.Chain, app.Mempool); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "network initialized")

	gui.GUI.Log("Main Loop")
	globals.MainEvents.BroadcastEvent("main", "initialized")

}
