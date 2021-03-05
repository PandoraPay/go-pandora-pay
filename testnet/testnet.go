package testnet

import (
	"pandora-pay/blockchain"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
)

type Testnet struct {
	wallet              *wallet.Wallet
	mempool             *mempool.MemPool
	chain               *blockchain.Blockchain
	transactionsBuilder *transactions_builder.TransactionsBuilder

	unstake  bool
	withdraw bool
}

func (testnet *Testnet) run() {

	for {

		blockHeight := <-testnet.chain.UpdateChannel
		if blockHeight > 50 {

			if !testnet.unstake {
				testnet.unstake = true
				tx, err := testnet.transactionsBuilder.CreateUnstakeTx(testnet.wa)
				if err != nil {
					gui.Error("Error creating testnet transaction", err)
					continue
				}
			}

		}

	}
}

func TestnetInit(wallet *wallet.Wallet, mempool *mempool.MemPool, chain *blockchain.Blockchain, transactionsBuilder *transactions_builder.TransactionsBuilder) (testnet *Testnet) {

	testnet = &Testnet{
		wallet:              wallet,
		mempool:             mempool,
		chain:               chain,
		transactionsBuilder: transactionsBuilder,
	}

	go testnet.run()

	return
}
