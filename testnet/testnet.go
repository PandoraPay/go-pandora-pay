package testnet

import (
	"encoding/hex"
	"github.com/tevino/abool"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
	wallet_address "pandora-pay/wallet/address"
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

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	tx, err := testnet.transactionsBuilder.CreateUnstakeTx(addr.AddressEncoded, 0, amount, 0, 0, true, []byte{}, true, true, true, true)
	if err != nil {
		return
	}

	gui.GUI.Info("Unstake transaction was created: " + hex.EncodeToString(tx.Bloom.Hash))

	return
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64) (err error) {

	dsts := []string{}
	dstsAmounts := []uint64{}
	dstsTokens := [][]byte{}
	for i := uint64(0); i < testnet.nodes; i++ {
		if uint64(testnet.wallet.GetAddressesCount()) <= i+1 {
			if _, err = testnet.wallet.AddNewAddress(true); err != nil {
				return
			}
		}

		var addr *wallet_address.WalletAddress
		addr, err = testnet.wallet.GetWalletAddress(int(i + 1))
		if err != nil {
			return
		}

		dsts = append(dsts, addr.AddressEncoded)
		dstsAmounts = append(dstsAmounts, config_stake.GetRequiredStake(blockHeight))
		dstsTokens = append(dstsTokens, config.NATIVE_TOKEN)
	}

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	tx, err := testnet.transactionsBuilder.CreateSimpleTx([]string{addr.AddressEncoded}, 0, []uint64{testnet.nodes * config_stake.GetRequiredStake(blockHeight)}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, 0, true, []byte{}, true, true, true, func(status string) {

	})
	if err != nil {
		return
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, hex.EncodeToString(tx.Bloom.Hash))

	return
}

func (testnet *Testnet) testnetCreateTransfers(blockHeight uint64) (err error) {

	dsts := []string{}
	dstsAmounts := []uint64{}
	dstsTokens := [][]byte{}

	count := rand.Intn(10) + 1
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

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	tx, err := testnet.transactionsBuilder.CreateSimpleTx([]string{addr.AddressEncoded}, 0, []uint64{sum}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, 0, true, []byte{}, true, true, true, func(status string) {

	})
	if err != nil {
		return
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, hex.EncodeToString(tx.Bloom.Hash))

	return
}

func (testnet *Testnet) run() {

	updateChannel := testnet.chain.UpdateNewChain.AddListener()
	defer testnet.chain.UpdateNewChain.RemoveChannel(updateChannel)

	creatingTransactions := abool.New()

	for {

		blockHeightReceived, ok := <-updateChannel
		if !ok {
			return
		}

		blockHeight := blockHeightReceived.(uint64)
		syncTime := testnet.chain.Sync.GetSyncTime()

		recovery.SafeGo(func() {

			gui.GUI.Log("UpdateNewChain received! 1")
			defer gui.GUI.Log("UpdateNewChain received! DONE")

			err := func() (err error) {

				if blockHeight == 20 {
					if err = testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*config_stake.GetRequiredStake(blockHeight)); err != nil {
						return
					}
				}
				if blockHeight == 30 {
					if err = testnet.testnetCreateTransfersNewWallets(blockHeight); err != nil {
						return
					}
				}

				if blockHeight >= 40 && syncTime != 0 {

					var addr *wallet_address.WalletAddress
					addr, err = testnet.wallet.GetWalletAddress(0)
					if err != nil {
						return
					}

					var balance, delegatedStakeAvailable, delegatedUnstakePending uint64
					var account *account.Account

					gui.GUI.Log("UpdateNewChain received! 2")

					if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

						accs := accounts.NewAccounts(reader)
						if account, err = accs.GetAccountEvenEmpty(addr.PublicKeyHash, blockHeight); err != nil {
							return
						}

						if account != nil {

							balance = account.GetAvailableBalance(config.NATIVE_TOKEN)

							delegatedStakeAvailable = account.GetDelegatedStakeAvailable()
							delegatedUnstakePending, _ = account.ComputeDelegatedUnstakePending()

						}

						return
					}); err != nil {
						return
					}

					if account != nil {

						if delegatedStakeAvailable > 0 && balance < delegatedStakeAvailable/4 && delegatedUnstakePending == 0 {
							if !testnet.mempool.ExistsTxSimpleVersion(addr.PublicKeyHash, transaction_simple.SCRIPT_UNSTAKE) {
								if err = testnet.testnetCreateUnstakeTx(blockHeight, delegatedStakeAvailable/2-balance); err != nil {
									//return
								}
							}
						} else {

							if creatingTransactions.IsNotSet() {
								creatingTransactions.Set()
								for {
									time.Sleep(time.Millisecond*time.Duration(rand.Intn(500)) + time.Millisecond*time.Duration(500))
									if testnet.mempool.CountInputTxs(addr.PublicKeyHash) < 20 {
										if err = testnet.testnetCreateTransfers(blockHeight); err != nil {
											//return
										}
									}
								}
							}
						}

					}

				}

				return
			}()

			if err != nil {
				gui.GUI.Error("Error creating testnet Tx", err)
				err = nil
			}

		})

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

	recovery.SafeGo(testnet.run)

	return
}
