package app

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/forging"
	"pandora-pay/mempool"
	"pandora-pay/network"
	"pandora-pay/settings"
	transactions_builder "pandora-pay/transactions-builder"
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
