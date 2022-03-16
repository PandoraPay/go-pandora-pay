package start

import (
	"context"
	"os"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/globals"
	balance_decoder "pandora-pay/cryptography/crypto/balance-decryptor"
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
	"runtime"
)

func _startMain() (err error) {

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

	if app.AddressBalanceDecryptor, err = address_balance_decryptor.NewAddressBalanceDecryptor(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "address balance decryptor validator initialized")

	if app.Mempool, err = mempool.CreateMempool(app.TxsValidator); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if app.Forging, err = forging.CreateForging(app.Mempool, app.AddressBalanceDecryptor); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if app.Chain, err = blockchain.CreateBlockchain(app.Mempool, app.TxsValidator); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	if app.Wallet, err = wallet.CreateWallet(app.Forging, app.Mempool, app.AddressBalanceDecryptor); err != nil {
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

	if runtime.GOARCH != "wasm" {
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			balance_decoder.BalanceDecryptor.SetTableSize(0, ctx, func(string) {})
		}()
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

	app.Forging.InitializeForging(app.TxsBuilder.CreateForgingTransactions, app.Chain.NextBlockCreatedCn, app.Chain.UpdateNewChainUpdate, app.Chain.ForgingSolutionCn)

	if config_forging.FORGING_ENABLED {
		app.Forging.StartForging()
	}

	app.Chain.InitForging()

	if globals.Arguments["--exit"] == true {
		os.Exit(1)
		return
	}

	if globals.Arguments["--new-devnet"] == true && runtime.GOARCH != "wasm" {
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
