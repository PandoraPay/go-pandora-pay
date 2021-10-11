package start

import (
	"os"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

func _startMain() (err error) {

	if globals.MainStarted {
		return
	}
	globals.MainStarted = true

	if globals.Arguments["--debugging"] == true {
		if err = debugging.Start(); err != nil {
			return
		}
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

	app.Forging.InitializeForging(app.Chain.NextBlockCreatedCn, app.Chain.UpdatePlainAccounts, app.Chain.ForgingSolutionCn)

	if app.Wallet, err = wallet.CreateWallet(app.Forging, app.Mempool); err != nil {
		return
	}
	if err = app.Wallet.ProcessWalletArguments(); err != nil {
		return
	}

	globals.MainEvents.BroadcastEvent("main", "wallet initialized")
	app.Wallet.InitializeWallet(app.Chain.UpdateAccounts, app.Chain.UpdatePlainAccounts)

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

	if globals.Arguments["--exit"] == true {
		os.Exit(1)
		return
	}

	if globals.Arguments["--new-devnet"] == true {
		myTestnet := testnet.TestnetInit(app.Wallet, app.Mempool, app.Chain, app.TransactionsBuilder)
		globals.Data["testnet"] = myTestnet
	}

	if app.Network, err = network.CreateNetwork(app.Settings, app.Chain, app.Mempool, app.Wallet, app.TransactionsBuilder); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "network initialized")

	gui.GUI.Log("Main Loop")
	globals.MainEvents.BroadcastEvent("main", "initialized")

	return
}

func startMain() {

	if err := _startMain(); err != nil {
		gui.GUI.Error(err)
	}

}
