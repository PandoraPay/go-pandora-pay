package start

import (
	"fmt"
	"os"
	"os/signal"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging_pprof"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"syscall"
)

func StartMainNow() (err error) {

	if globals.MainStarted {
		return
	}
	globals.MainStarted = true

	if globals.Arguments["--pprof"] == true {
		if err = debugging_pprof.Start(); err != nil {
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

	if app.TxsValidator, err = txs_validator.NewTxsValidator(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "txs validator initialized")

	globals.MainEvents.BroadcastEvent("main", "address balance decryptor validator initialized")

	if app.Mempool, err = mempool.CreateMempool(app.TxsValidator); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if app.Forging, err = forging.CreateForging(app.Mempool); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if app.Chain, err = blockchain.CreateBlockchain(app.Mempool, app.TxsValidator); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	if app.Wallet, err = wallet.CreateWallet(app.Forging, app.Mempool); err != nil {
		return
	}
	if err = app.Wallet.ProcessWalletArguments(); err != nil {
		return
	}

	globals.MainEvents.BroadcastEvent("main", "wallet initialized")

	if err = genesis.GenesisInit(app.Wallet.GetFirstAddressForDevnetGenesisAirdrop); err != nil {
		return
	}
	if err = app.Chain.InitializeChain(); err != nil {
		return
	}

	app.Wallet.InitializeWallet(app.Chain.UpdateNewChainUpdate)
	if err = app.Wallet.StartWallet(); err != nil {
		return
	}

	if app.Settings, err = settings.SettingsInit(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "settings initialized")

	app.TxsBuilder = txs_builder.TxsBuilderInit(app.Wallet, app.Mempool, app.TxsValidator)
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	app.Forging.InitializeForging(app.Chain.NextBlockCreatedCn, app.Chain.UpdateNewChainUpdate, app.Chain.ForgingSolutionCn)

	if config_forging.FORGING_ENABLED {
		app.Forging.StartForging()
		if err = app.Wallet.InitForgingWallet(); err != nil {
			return
		}
	}

	app.Chain.InitForging()

	if globals.Arguments["--exit"] == true {
		os.Exit(1)
		return
	}

	if globals.Arguments["--run-testnet-script"] == true {
		myTestnet := testnet.TestnetInit(app.Wallet, app.Mempool, app.Chain, app.TxsBuilder)
		globals.Data["testnet"] = myTestnet
	}

	if app.Network, err = network.NewNetwork(app.Settings, app.Chain, app.Mempool, app.Wallet, app.TxsValidator, app.TxsBuilder); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "network initialized")

	gui.GUI.Log("Main Loop")
	globals.MainEvents.BroadcastEvent("main", "initialized")

	return
}

func InitMain(ready func()) {
	var err error

	argv := os.Args[1:]
	if err = arguments.InitArguments(argv); err != nil {
		saveError(err)
	}

	if err = config.InitConfig(); err != nil {
		saveError(err)
	}
	globals.MainEvents.BroadcastEvent("main", "config initialized")

	startMain()

	if ready != nil {
		ready()
	}

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	fmt.Println("Shutting down")
}
