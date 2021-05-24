package main

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/config/arguments"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/settings"
	"pandora-pay/store"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"strings"
	"syscall/js"
)

func main() {

	var err error
	var mySettings *settings.Settings
	var myWallet *wallet.Wallet
	var myChain *blockchain.Blockchain
	var myForging *forging.Forging
	var myMempool *mempool.Mempool

	config.StartConfig()

	args := []string{}
	jsConfig := js.Global().Get("PandoraPayConfig")
	if jsConfig.Truthy() {
		if jsConfig.Type() != js.TypeString {
			panic("PandoraPayConfig must be a string")
		}
		args = strings.Split(jsConfig.String(), " ")
	}

	if err = arguments.InitArguments(args); err != nil {
		panic(err)
	}

	defer func() {
		err := recover()
		if err != nil && gui.GUI != nil {
			gui.GUI.Error(err)
			gui.GUI.Close()
		}
	}()

	if err = gui.InitGUI(); err != nil {
		panic(err)
	}

	for i, arg := range args {
		gui.GUI.Log("Argument", i, arg)
	}

	if err = config.InitConfig(); err != nil {
		panic(err)
	}

	if err = store.InitDB(); err != nil {
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

	gui.GUI.Log("wallet")

	if mySettings, err = settings.SettingsInit(); err != nil {
		panic(err)
	}
	globals.Data["settings"] = mySettings

	if myChain, err = blockchain.BlockchainInit(nil, myWallet, myMempool); err != nil {
		panic(err)
	}
	globals.Data["chain"] = myChain

	myTransactionsBuilder := transactions_builder.TransactionsBuilderInit(myWallet, myMempool, myChain)
	globals.Data["transactionsBuilder"] = myTransactionsBuilder

}
