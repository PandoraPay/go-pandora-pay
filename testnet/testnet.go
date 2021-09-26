package testnet

import (
	"encoding/hex"
	"github.com/tevino/abool"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/transactions-builder/wizard"
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

func (testnet *Testnet) testnetCreateClaimTx(reg bool, amount uint64) (tx *transaction.Transaction, err error) {

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	if tx, err = testnet.transactionsBuilder.CreateClaimTx(addr.AddressEncoded, 0, []uint64{amount}, []string{addr.AddressRegistrationEncoded}, &wizard.TransactionsWizardData{nil, false},
		&wizard.TransactionsWizardFee{0, 0, 0, true}, true, true, true, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Claim tx was created: " + hex.EncodeToString(tx.Bloom.Hash))
	return
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight uint64, amount uint64) (tx *transaction.Transaction, err error) {

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	if tx, err = testnet.transactionsBuilder.CreateUnstakeTx(addr.AddressEncoded, 0, amount, &wizard.TransactionsWizardData{nil, false}, &wizard.TransactionsWizardFee{0, 0, 0, true}, true, true, true, func(string) {}); tx != nil {
		return nil, err
	}

	gui.GUI.Info("Unstake tx was created: " + hex.EncodeToString(tx.Bloom.Hash))
	return
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64) (tx *transaction.Transaction, err error) {

	from := []string{}
	dsts := []string{}
	dstsAmounts, burn := []uint64{}, []uint64{}
	dstsTokens := [][]byte{}
	data := [][]byte{}
	ringMembers := [][]string{}
	fees := []*wizard.TransactionsWizardFee{}

	for i := uint64(0); i < testnet.nodes; i++ {

		if uint64(testnet.wallet.GetAddressesCount()) <= i+1 {
			if _, err = testnet.wallet.AddNewAddress(true); err != nil {
				return
			}
		}

		var addr *wallet_address.WalletAddress

		if addr, err = testnet.wallet.GetWalletAddress(0); err != nil {
			return
		}
		from = append(from, addr.AddressEncoded)

		if addr, err = testnet.wallet.GetWalletAddress(int(i + 1)); err != nil {
			return
		}

		token := config.NATIVE_TOKEN

		dsts = append(dsts, addr.AddressRegistrationEncoded)
		dstsAmounts = append(dstsAmounts, config_stake.GetRequiredStake(blockHeight))
		dstsTokens = append(dstsTokens, token)
		burn = append(burn, 0)

		var ring []string
		if ring, err = testnet.transactionsBuilder.CreateZetherRing(from[i], addr.AddressEncoded, token, -1, -1); err != nil {
			return
		}
		ringMembers = append(ringMembers, ring)

		data = append(data, []byte{})
		fees = append(fees, &wizard.TransactionsWizardFee{0, 0, 0, true})
	}

	if tx, err = testnet.transactionsBuilder.CreateZetherTx(from, dstsTokens, dstsAmounts, dsts, burn, ringMembers, data, fees, true, true, true, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, hex.EncodeToString(tx.Bloom.Hash))
	return
}

func (testnet *Testnet) testnetCreateTransfers(blockHeight uint64) (tx *transaction.Transaction, err error) {

	amount := uint64(rand.Int63n(6))
	burn := uint64(0)

	privateKey := addresses.GenerateNewPrivateKey()

	addr, err := privateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	if err != nil {
		return
	}

	dst := addr.EncodeAddr()

	walletAddr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	data := []byte{}
	fee := &wizard.TransactionsWizardFee{0, 0, 0, true}

	ringMembers, err := testnet.transactionsBuilder.CreateZetherRing(walletAddr.AddressEncoded, dst, config.NATIVE_TOKEN, -1, -1)
	if err != nil {
		return
	}

	if tx, err = testnet.transactionsBuilder.CreateZetherTx([]string{walletAddr.AddressEncoded}, [][]byte{config.NATIVE_TOKEN}, []uint64{amount}, []string{dst}, []uint64{burn}, [][]string{ringMembers}, [][]byte{data}, []*wizard.TransactionsWizardFee{fee}, true, true, true, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).Height, hex.EncodeToString(tx.Bloom.Hash))
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
					if _, err = testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*config_stake.GetRequiredStake(blockHeight)); err != nil {
						return
					}
				}
				if blockHeight == 100 {
					if _, err = testnet.testnetCreateTransfersNewWallets(blockHeight); err != nil {
						return
					}
				}

				if blockHeight >= 40 && syncTime != 0 {

					var addr *wallet_address.WalletAddress
					addr, err = testnet.wallet.GetWalletAddress(0)
					if err != nil {
						return
					}

					publicKey := addr.PublicKey

					var delegatedStakeAvailable, delegatedUnstakePending, claimable uint64
					var balanceHomo *crypto.ElGamal

					var acc *account.Account
					var plainAcc *plain_account.PlainAccount
					var reg bool

					gui.GUI.Log("UpdateNewChain received! 2")

					if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

						accsCollection := accounts.NewAccountsCollection(reader)
						regs := registrations.NewRegistrations(reader)

						accs, err := accsCollection.GetMap(config.NATIVE_TOKEN)
						if err != nil {
							return
						}
						if acc, err = accs.GetAccount(publicKey); err != nil {
							return
						}

						if reg, err = regs.Exists(string(publicKey)); err != nil {
							return
						}

						plainAccs := plain_accounts.NewPlainAccounts(reader)
						if plainAcc, err = plainAccs.GetPlainAccount(publicKey, blockHeight); err != nil {
							return
						}

						if acc != nil {
							balanceHomo = acc.GetBalance()
						}

						if plainAcc != nil {
							delegatedStakeAvailable = plainAcc.GetDelegatedStakeAvailable()
							delegatedUnstakePending, _ = plainAcc.ComputeDelegatedUnstakePending()
							claimable = plainAcc.Claimable
						}

						return
					}); err != nil {
						return
					}

					if acc != nil || plainAcc != nil {

						var balance uint64
						if acc != nil {
							if balance, err = testnet.wallet.DecodeBalanceByPublicKey(publicKey, balanceHomo, config.NATIVE_TOKEN, false); err != nil {
								return
							}
						}

						if claimable > 0 {
							if !testnet.mempool.ExistsTxSimpleVersion(addr.PublicKey, transaction_simple.SCRIPT_CLAIM) {
								if _, err = testnet.testnetCreateClaimTx(reg, claimable); err != nil {
									return
								}
							}
						} else if delegatedStakeAvailable > 0 && balance < delegatedStakeAvailable/4 && delegatedUnstakePending == 0 {
							if !testnet.mempool.ExistsTxSimpleVersion(addr.PublicKey, transaction_simple.SCRIPT_UNSTAKE) {
								if _, err = testnet.testnetCreateUnstakeTx(blockHeight, delegatedStakeAvailable/2-balance); err != nil {
									return
								}
							}
						} else {

							if creatingTransactions.IsNotSet() {
								creatingTransactions.Set()
								for {
									time.Sleep(time.Millisecond*time.Duration(rand.Intn(500)) + time.Millisecond*time.Duration(500))
									if testnet.mempool.CountInputTxs(addr.PublicKey) < 20 {
										if _, err = testnet.testnetCreateTransfers(blockHeight); err != nil {
											return
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
