package app

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	"pandora-pay/store"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

var (
	Settings            *settings.Settings
	Wallet              *wallet.Wallet
	Forging             *forging.Forging
	Mempool             *mempool.Mempool
	Chain               *blockchain.Blockchain
	Network             *network.Network
	TransactionsBuilder *transactions_builder.TransactionsBuilder
)

func Close() {
	store.DBClose()
	gui.GUI.Close()
	Forging.Close()
	Chain.Close()
	Wallet.Close()
}
