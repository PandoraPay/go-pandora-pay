package testnet

import (
	"encoding/hex"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/config"
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

	nodes uint64
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight, amount uint64) {

	tx := testnet.transactionsBuilder.CreateUnstakeTx(testnet.wallet.Addresses[0].AddressEncoded, amount, -1, []byte{}, true)
	hash := tx.ComputeHash()
	gui.Info("Unstake transaction was created: " + hex.EncodeToString(hash[:]))

	result := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if !result {
		panic("transaction was not inserted in mempool")
	}
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64) {
	dsts := make([]string, 0)
	dstsAmounts := make([]uint64, 0)
	dstsTokens := make([][]byte, 0)
	for i := uint64(0); i < testnet.nodes; i++ {
		if uint64(len(testnet.wallet.Addresses)) <= i+1 {
			testnet.wallet.AddNewAddress()
		}
		dsts = append(dsts, testnet.wallet.Addresses[i+1].AddressEncoded)
		dstsAmounts = append(dstsAmounts, stake.GetRequiredStake(blockHeight))
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
	}

	tx := testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].AddressEncoded}, []uint64{testnet.nodes * stake.GetRequiredStake(blockHeight)}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
	hash := tx.ComputeHash()
	gui.Info("Create Transfers transaction was created: " + hex.EncodeToString(hash[:]))

	result := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if !result {
		panic("transaction was not inserted in mempool")
	}
}

func (testnet *Testnet) testnetCreateTransfers(blockHeight uint64) {
	dsts := make([]string, 0)
	dstsAmounts := make([]uint64, 0)
	dstsTokens := make([][]byte, 0)

	count := rand.Intn(19) + 1
	sum := uint64(0)
	for i := 0; i < count; i++ {
		privateKey := addresses.GenerateNewPrivateKey()
		addr := privateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
		dsts = append(dsts, addr.EncodeAddr())
		amount := uint64(rand.Int63n(10))
		dstsAmounts = append(dstsAmounts, amount)
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
		sum += amount
	}

	tx := testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].AddressEncoded}, []uint64{sum}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
	hash := tx.ComputeHash()
	gui.Info("Create Transfers transaction was created: " + hex.EncodeToString(hash[:]))

	result := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if !result {
		panic("transaction was not inserted in mempool")
	}
}

func (testnet *Testnet) run() {

	for {

		blockHeight := <-testnet.chain.UpdateChannel

		func() {

			defer func() {
				if err := helpers.ConvertRecoverError(recover()); err != nil {
					gui.Error("Error creating testnet Tx", err)
				}
			}()

			if blockHeight == 30 {
				testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*stake.GetRequiredStake(blockHeight))
			}
			if blockHeight == 50 {
				testnet.testnetCreateTransfersNewWallets(blockHeight)
			}

			if blockHeight >= 60 {
				if blockHeight%20 == 0 {
					testnet.testnetCreateUnstakeTx(blockHeight, 20*20*10)
				} else {
					testnet.testnetCreateTransfers(blockHeight)
				}
			}

		}()

	}
}

func TestnetInit(wallet *wallet.Wallet, mempool *mempool.MemPool, chain *blockchain.Blockchain, transactionsBuilder *transactions_builder.TransactionsBuilder) (testnet *Testnet) {

	testnet = &Testnet{
		wallet:              wallet,
		mempool:             mempool,
		chain:               chain,
		transactionsBuilder: transactionsBuilder,
		nodes:               uint64(config.CPU_THREADS),
	}

	go testnet.run()

	return
}
