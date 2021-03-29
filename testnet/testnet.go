package testnet

import (
	"encoding/hex"
	"errors"
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
	mempool             *mempool.Mempool
	chain               *blockchain.Blockchain
	transactionsBuilder *transactions_builder.TransactionsBuilder

	nodes uint64
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight, amount uint64) (err error) {

	tx, err := testnet.transactionsBuilder.CreateUnstakeTx(testnet.wallet.Addresses[0].AddressEncoded, amount, -1, []byte{}, true)
	if err != nil {
		return
	}

	gui.Info("Unstake transaction was created: " + hex.EncodeToString(tx.Bloom.Hash))

	result, err := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if err != nil {
		return
	}
	if !result {
		return errors.New("transaction was not inserted in mempool")
	}
	return
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64) (err error) {
	dsts := []string{}
	dstsAmounts := []uint64{}
	dstsTokens := [][]byte{}
	for i := uint64(0); i < testnet.nodes; i++ {
		if uint64(len(testnet.wallet.Addresses)) <= i+1 {
			testnet.wallet.AddNewAddress()
		}
		dsts = append(dsts, testnet.wallet.Addresses[i+1].AddressEncoded)
		dstsAmounts = append(dstsAmounts, stake.GetRequiredStake(blockHeight))
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
	}

	tx, err := testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].AddressEncoded}, []uint64{testnet.nodes * stake.GetRequiredStake(blockHeight)}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
	if err != nil {
		return
	}

	gui.Info("Create Transfers transaction was created: " + hex.EncodeToString(tx.Bloom.Hash))

	result, err := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if err != nil {
		return
	}
	if !result {
		return errors.New("transaction was not inserted in mempool")
	}
	return
}

func (testnet *Testnet) testnetCreateTransfers(blockHeight uint64) error {
	dsts := []string{}
	dstsAmounts := []uint64{}
	dstsTokens := [][]byte{}

	count := rand.Intn(19) + 1
	sum := uint64(0)
	for i := 0; i < count; i++ {
		privateKey := addresses.GenerateNewPrivateKey()
		addr, err := privateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
		if err != nil {
			return err
		}
		dsts = append(dsts, addr.EncodeAddr())
		amount := uint64(rand.Int63n(6))
		dstsAmounts = append(dstsAmounts, amount)
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
		sum += amount
	}

	tx, err := testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].AddressEncoded}, []uint64{sum}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
	if err != nil {
		return nil
	}

	gui.Info("Create Transfers transaction was created: " + hex.EncodeToString(tx.Bloom.Hash))

	result, err := testnet.mempool.AddTxToMemPool(tx, blockHeight, true)
	if err != nil {
		return err
	}
	if !result {
		return errors.New("transaction was not inserted in mempool")
	}
	return nil
}

func (testnet *Testnet) run() {

	var err error
	for {

		blockHeight := <-testnet.chain.UpdateChannel

		func() {

			if blockHeight == 30 {
				err = testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*stake.GetRequiredStake(blockHeight))
			}
			if blockHeight == 50 {
				err = testnet.testnetCreateTransfersNewWallets(blockHeight)
			}

			if blockHeight >= 60 {
				if blockHeight%20 == 0 {
					err = testnet.testnetCreateUnstakeTx(blockHeight, 20*20*20*5)
				} else {
					for i := 0; i < 20; i++ {
						err = testnet.testnetCreateTransfers(blockHeight)
					}
				}
			}

			if err != nil {
				gui.Error("Error creating testnet Tx", err)
				err = nil
			}

		}()

	}
}

func TestnetInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain, transactionsBuilder *transactions_builder.TransactionsBuilder) (testnet *Testnet) {

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
