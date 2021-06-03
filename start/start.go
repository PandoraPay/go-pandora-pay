package start

import (
	"fmt"
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
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
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

	if myMempool, err = mempool.InitMemPool(); err != nil {
		panic(err)
	}
	globals.Data["mempool"] = myMempool
	globals.MainEvents.BroadcastEvent("main", "mempool initialized")

	if myChain, err = blockchain.BlockchainInit(myMempool); err != nil {
		panic(err)
	}
	globals.Data["chain"] = myChain
	globals.MainEvents.BroadcastEvent("main", "blockchain initialized")

	if myForging, err = forging.ForgingInit(myMempool, myChain.NextBlockCreatedCn, myChain.UpdateAccounts, myChain.ForgingSolutionCn); err != nil {
		panic(err)
	}
	globals.Data["forging"] = myForging
	globals.MainEvents.BroadcastEvent("main", "forging initialized")

	if myWallet, err = wallet.WalletInit(myForging, myMempool, myChain.UpdateAccounts); err != nil {
		panic(err)
	}
	globals.Data["wallet"] = myWallet
	globals.MainEvents.BroadcastEvent("main", "wallet initialized")

	if err = myWallet.InitializeWallet(); err != nil {
		return
	}
	myChain.InitForging()
	myForging.StartForging()

	if mySettings, err = settings.SettingsInit(); err != nil {
		panic(err)
	}
	globals.Data["settings"] = mySettings
	globals.MainEvents.BroadcastEvent("main", "settings initialized")

	if err = genesis.GenesisInit(myWallet); err != nil {
		return
	}
	globals.MainEvents.BroadcastEvent("main", "genesis created")

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

func CloseMain() {
	gui.GUI.Close()
	myForging.Close()
	myChain.Close()
	myWallet.Close()
	store.DBClose()
}
