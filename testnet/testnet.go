package testnet

import (
	"encoding/hex"
	"pandora-pay/blockchain"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
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

	nodes uint64
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight uint64) (err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()

	tx := testnet.transactionsBuilder.CreateUnstakeTx(testnet.wallet.Addresses[0].AddressEncoded, testnet.nodes*stake.GetRequiredStake(blockHeight), -1, []byte{}, true)
	hash := tx.ComputeHash()
	gui.Info("Unstake transaction was created: " + hex.EncodeToString(hash[:]))

	result := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if !result {
		panic("transaction was not isnerted in mempool")
	}

	return
}

func (testnet *Testnet) run() {

	for {

		blockHeight := <-testnet.chain.UpdateChannel
		if blockHeight == 40 {

			testnet.unstake = true
			if err := testnet.testnetCreateUnstakeTx(blockHeight); err != nil {
				gui.Error("Error creating unstake Tx", err)
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
		nodes:               uint64(5),
	}

	go testnet.run()

	return
}
