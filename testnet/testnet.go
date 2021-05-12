package testnet

import (
	"encoding/hex"
	"errors"
	bolt "go.etcd.io/bbolt"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/config"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/store"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	"time"
)

type Testnet struct {
	wallet              *wallet.Wallet
	mempool             *mempool.Mempool
	chain               *blockchain.Blockchain
	transactionsBuilder *transactions_builder.TransactionsBuilder
	nodes               uint64
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight uint64, amount uint64) (err error) {

	tx, err := testnet.transactionsBuilder.CreateUnstakeTx(testnet.wallet.Addresses[0].GetAddressEncoded(), 0, amount, -1, []byte{}, true)
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
			if _, err = testnet.wallet.AddNewAddress(); err != nil {
				return
			}
		}
		dsts = append(dsts, testnet.wallet.Addresses[i+1].GetAddressEncoded())
		dstsAmounts = append(dstsAmounts, stake.GetRequiredStake(blockHeight))
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
	}

	tx, err := testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].GetAddressEncoded()}, 0, []uint64{testnet.nodes * stake.GetRequiredStake(blockHeight)}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
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

func (testnet *Testnet) testnetCreateTransfers(blockHeight uint64) (err error) {
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

	var tx *transaction.Transaction
	if tx, err = testnet.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].GetAddressEncoded()}, 0, []uint64{sum}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, -1, []byte{}); err != nil {
		return
	}

	gui.Info("Create Transfers transaction was created: " + hex.EncodeToString(tx.Bloom.Hash))

	var result bool
	if result, err = testnet.mempool.AddTxToMemPool(tx, blockHeight, true); err != nil {
		return
	}
	if !result {
		return errors.New("transaction was not inserted in mempool")
	}

	return
}

func (testnet *Testnet) run() {

	updateChannel := testnet.chain.UpdateMulticast.AddListener()
	for {

		blockHeightReceived, ok := <-updateChannel
		if !ok {
			return
		}

		blockHeight := blockHeightReceived.(uint64)
		syncTime := testnet.chain.Sync.GetSyncTime()

		err := func() (err error) {

			if blockHeight == 30 {
				if err = testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*stake.GetRequiredStake(blockHeight)); err != nil {
					return
				}
			}
			if blockHeight == 50 {
				if err = testnet.testnetCreateTransfersNewWallets(blockHeight); err != nil {
					return
				}
			}

			if blockHeight >= 60 && syncTime != 0 {

				if err = store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

					accs := accounts.NewAccounts(boltTx)
					var account *account.Account
					if account, err = accs.GetAccountEvenEmpty(testnet.wallet.Addresses[0].GetPublicKeyHash(), blockHeight); err != nil {
						return
					}

					if account != nil {

						var balance, delegatedStakeAvailable uint64
						if balance, err = account.GetAvailableBalance(config.NATIVE_TOKEN); err != nil {
							return
						}
						delegatedStakeAvailable = account.GetDelegatedStakeAvailable()

						if delegatedStakeAvailable > 0 && balance < delegatedStakeAvailable/4 {
							if !testnet.mempool.ExistsTxSimpleVersion(testnet.wallet.Addresses[0].GetPublicKeyHash(), transaction_simple.TxSimpleScriptUnstake) {
								if err = testnet.testnetCreateUnstakeTx(blockHeight, delegatedStakeAvailable/2-balance); err != nil {
									return
								}
							}
						} else {
							if testnet.mempool.CountInputTxs(testnet.wallet.Addresses[0].GetPublicKeyHash()) < 100 {
								for i := 0; i < 20; i++ {
									err = testnet.testnetCreateTransfers(blockHeight)
									time.Sleep(10 * time.Millisecond)
								}
							}
						}

					}

					return
				}); err != nil {
					return
				}

			}

			return
		}()

		if err != nil {
			gui.Error("Error creating testnet Tx", err)
			err = nil
		}

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
