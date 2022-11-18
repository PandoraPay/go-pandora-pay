package start

import (
	"context"
	"fmt"
	"math"
	"os"
	"os/signal"
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/app"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/config_forging"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography/crypto/balance_decryptor"
	"pandora-pay/gui"
	"pandora-pay/helpers/debugging_pprof"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/network/network_config"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/testnet"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"runtime"
	"strconv"
	"syscall"
)

func StartMainNow() (err error) {

	if !globals.MainStarted.CompareAndSwap(false, true) {
		return
	}

	arguments.VERSION_STRING = config.VERSION_STRING
	if arguments.Arguments["--pprof"] == true {
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

	if err = txs_validator.NewTxsValidator(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "txs validator initialized")

	if app.AddressBalanceDecryptor, err = address_balance_decryptor.NewAddressBalanceDecryptor(runtime.GOARCH != "wasm"); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "address balance decryptor validator initialized")

	if app.Mempool, err = mempool.CreateMempool(); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if app.Forging, err = forging.CreateForging(app.Mempool, app.AddressBalanceDecryptor); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if app.Chain, err = blockchain.CreateBlockchain(app.Mempool); err != nil {
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

	if runtime.GOARCH != "wasm" && arguments.Arguments["--balance-decryptor-disable-init"] == false {
		tableSize := 0
		if arguments.Arguments["--balance-decryptor-table-size"] != nil {
			if tableSize, err = strconv.Atoi(arguments.Arguments["--balance-decryptor-table-size"].(string)); err != nil {
				return
			}
			tableSize = 1 << tableSize
		}
		go func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			gui.GUI.Info2Update("Decryptor", "Init... "+strconv.Itoa(int(math.Log2(float64(tableSize)))))
			balance_decryptor.BalanceDecryptor.SetTableSize(tableSize, ctx, func(string) {})
			gui.GUI.Info2Update("Decryptor", "Ready "+strconv.Itoa(int(math.Log2(float64(tableSize)))))
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

	if err = txs_builder.TxsBuilderInit(app.Wallet, app.Mempool); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "transactions builder initialized")

	app.Forging.InitializeForging(txs_builder.TxsBuilder.CreateForgingTransactions, app.Chain.NextBlockCreatedCn, app.Chain.UpdateNewChainUpdate, app.Chain.ForgingSolutionCn)

	if config_forging.FORGING_ENABLED {
		app.Forging.StartForging()
	}

	app.Chain.InitForging()

	if arguments.Arguments["--exit"] == true {
		os.Exit(1)
		return
	}

	if arguments.Arguments["--run-testnet-script"] == true {
		if err = testnet.TestnetInit(app.Wallet, app.Mempool, app.Chain); err != nil {
			return
		}
	}

	if err = network.NewNetwork(app.Settings, app.Chain, app.Mempool, app.Wallet); err != nil {
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
	globals.MainEvents.BroadcastEvent("main", "arguments initialized")

	if err = config.InitConfig(); err != nil {
		saveError(err)
	}
	globals.MainEvents.BroadcastEvent("main", "config initialized")
	if err = network_config.InitConfig(); err != nil {
		return
	}

	startMain()

	if ready != nil {
		ready()
	}

	exitSignal := make(chan os.Signal, 10)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)
	<-exitSignal

	fmt.Println("Shutting down")
	app.Close()
}
