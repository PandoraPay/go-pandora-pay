package start

import (
	"context"
	"os"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config/globals"
	balance_decoder "pandora-pay/cryptography/crypto/balance-decoder"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"runtime"
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

	if app.TxsValidator, err = txs_validator.NewTxsValidator(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "txs validator initialized")

	if app.Mempool, err = mempool.CreateMempool(); err != nil {
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

	if runtime.GOARCH != "wasm" {
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			balance_decoder.BalanceDecoder.SetTableSize(0, ctx, func(string) {})
		}()
	}

	app.TxsBuilder = txs_builder.TxsBuilderInit(app.Wallet, app.Mempool, app.Chain)
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	if globals.Arguments["--exit"] == true {
		os.Exit(1)
		return
	}

	if globals.Arguments["--new-devnet"] == true {
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

func startMain() {

	if err := _startMain(); err != nil {
		gui.GUI.Error(err)
	}

}
