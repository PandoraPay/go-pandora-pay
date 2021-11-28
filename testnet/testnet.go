package testnet

import (
	"context"
	"encoding/hex"
	"github.com/tevino/abool"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/transactions_builder"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
	"sync/atomic"
	"time"
)

type Testnet struct {
	wallet              *wallet.Wallet
	mempool             *mempool.Mempool
	chain               *blockchain.Blockchain
	transactionsBuilder *transactions_builder.TransactionsBuilder
	nodes               uint64
}

func (testnet *Testnet) testnetCreateClaimTx(dstAddressWalletIndex int, amount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	select {
	case <-ctx.Done():
		return
	default:
	}

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	dstAddr, err := testnet.wallet.GetWalletAddress(dstAddressWalletIndex)
	if err != nil {
		return
	}

	from := []string{""}
	dsts := []string{dstAddr.AddressRegistrationEncoded}
	dstsAmounts, burn := []uint64{amount}, []uint64{0}
	dstsAssets := [][]byte{config_coins.NATIVE_ASSET_FULL}
	data := []*wizard.WizardTransactionData{{[]byte{}, false}}
	fees := []*wizard.WizardZetherTransactionFee{{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}}

	if tx, err = testnet.transactionsBuilder.CreateZetherTx([]wizard.WizardZetherPayloadExtra{&wizard.WizardZetherPayloadExtraClaim{DelegatePrivateKey: addr.PrivateKey.Key}}, from, dstsAssets, dstsAmounts, dsts, burn, []*transactions_builder.ZetherRingConfiguration{{-1, -1}}, data, fees, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Claim Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight uint64, amount uint64) (tx *transaction.Transaction, err error) {

	addr, err := testnet.wallet.GetWalletAddress(0)
	if err != nil {
		return
	}

	if tx, err = testnet.transactionsBuilder.CreateSimpleTx(addr.AddressEncoded, 0, &wizard.WizardTxSimpleExtraUnstake{Amount: amount}, &wizard.WizardTransactionData{nil, false}, &wizard.WizardTransactionFee{0, 0, 0, true}, false, true, true, true, false, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Unstake tx was created: " + hex.EncodeToString(tx.Bloom.Hash))
	return
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	from := []string{}
	dsts := []string{}
	dstsAmounts, burn := []uint64{}, []uint64{}
	dstsAssets := [][]byte{}
	data := []*wizard.WizardTransactionData{}
	ringsConfigurations := []*transactions_builder.ZetherRingConfiguration{}
	fees := []*wizard.WizardZetherTransactionFee{}
	payloadsExtra := []wizard.WizardZetherPayloadExtra{}

	for i := uint64(0); i < testnet.nodes; i++ {

		var addr *wallet_address.WalletAddress

		if addr, err = testnet.wallet.GetWalletAddress(1); err != nil {
			return
		}
		from = append(from, addr.AddressEncoded)

		if addr, err = testnet.wallet.GetWalletAddress(int(i + 1)); err != nil {
			return
		}

		asset := config_coins.NATIVE_ASSET_FULL

		dsts = append(dsts, addr.AddressRegistrationEncoded)
		dstsAmounts = append(dstsAmounts, config_stake.GetRequiredStake(blockHeight))
		dstsAssets = append(dstsAssets, asset)
		burn = append(burn, 0)

		ringsConfigurations = append(ringsConfigurations, &transactions_builder.ZetherRingConfiguration{-1, -1})

		data = append(data, &wizard.WizardTransactionData{[]byte{}, false})
		fees = append(fees, &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0})
		payloadsExtra = append(payloadsExtra, nil)
	}

	if tx, err = testnet.transactionsBuilder.CreateZetherTx(payloadsExtra, from, dstsAssets, dstsAmounts, dsts, burn, ringsConfigurations, data, fees, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) testnetCreateTransfers(srcAddressWalletIndex int, ctx context.Context) (tx *transaction.Transaction, err error) {

	select {
	case <-ctx.Done():
		return
	default:
	}

	srcAddr, err := testnet.wallet.GetWalletAddress(srcAddressWalletIndex)
	if err != nil {
		return
	}

	amount := uint64(rand.Int63n(6))
	burn := uint64(0)

	privateKey := addresses.GenerateNewPrivateKey()

	addr, err := privateKey.GenerateAddress(true, 0, helpers.EmptyBytes(0))
	if err != nil {
		return
	}

	dst := addr.EncodeAddr()

	data := &wizard.WizardTransactionData{nil, false}
	fees := []*wizard.WizardZetherTransactionFee{{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}}

	if tx, err = testnet.transactionsBuilder.CreateZetherTx([]wizard.WizardZetherPayloadExtra{nil}, []string{srcAddr.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{amount}, []string{dst}, []uint64{burn}, []*transactions_builder.ZetherRingConfiguration{{-1, -1}}, []*wizard.WizardTransactionData{data}, fees, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) run() {

	updateChannel := testnet.chain.UpdateNewChain.AddListener()
	defer testnet.chain.UpdateNewChain.RemoveChannel(updateChannel)

	creatingTransactions := abool.New()
	unstakesCount := int32(0)

	for i := uint64(0); i < testnet.nodes; i++ {
		if uint64(testnet.wallet.GetAddressesCount()) <= i+1 {
			if _, err := testnet.wallet.AddNewAddress(true); err != nil {
				return
			}
		}
	}

	var oldCancel context.CancelFunc

	for {

		blockHeightReceived, ok := <-updateChannel
		if !ok {
			return
		}

		if oldCancel != nil {
			oldCancel()
		}
		ctx, cancel := context.WithCancel(context.Background())
		oldCancel = cancel

		blockHeight := blockHeightReceived.(uint64)
		syncTime := testnet.chain.Sync.GetSyncTime()

		recovery.SafeGo(func() {

			gui.GUI.Log("UpdateNewChain received! 1")
			defer gui.GUI.Log("UpdateNewChain received! DONE")

			if err := func() (err error) {

				if blockHeight == 20 {
					if _, err = testnet.testnetCreateUnstakeTx(blockHeight, testnet.nodes*config_stake.GetRequiredStake(blockHeight)); err != nil {
						return
					}
				}
				if blockHeight == 100 {
					if _, err = testnet.testnetCreateTransfersNewWallets(blockHeight, ctx); err != nil {
						return
					}
				}

				if blockHeight >= 40 && syncTime != 0 {

					var addr *wallet_address.WalletAddress
					addr, _ = testnet.wallet.GetWalletAddress(0)

					var delegatedStakeAvailable, delegatedUnstakePending, unclaimed uint64

					var plainAcc *plain_account.PlainAccount

					gui.GUI.Log("UpdateNewChain received! 2")

					if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

						plainAccs := plain_accounts.NewPlainAccounts(reader)
						if plainAcc, err = plainAccs.GetPlainAccount(addr.PublicKey, blockHeight); err != nil {
							return
						}

						if plainAcc != nil {
							delegatedStakeAvailable = plainAcc.DelegatedStake.GetDelegatedStakeAvailable()
							delegatedUnstakePending, _ = plainAcc.DelegatedStake.ComputeDelegatedUnstakePending()
							unclaimed = plainAcc.Unclaimed
						}

						return
					}); err != nil {
						return
					}

					if plainAcc != nil {

						if creatingTransactions.IsNotSet() {

							creatingTransactions.Set()
							defer creatingTransactions.UnSet()

							if unclaimed > config_coins.ConvertToUnitsUint64Forced(40) {

								unclaimed -= config_coins.ConvertToUnitsUint64Forced(30)

								if !testnet.mempool.ExistsTxZetherVersion(addr.PublicKey, transaction_zether_payload.SCRIPT_CLAIM) {
									testnet.testnetCreateClaimTx(1, unclaimed/5, ctx)
									testnet.testnetCreateClaimTx(2, unclaimed/5, ctx)
									testnet.testnetCreateClaimTx(3, unclaimed/5, ctx)
									testnet.testnetCreateClaimTx(4, unclaimed/5, ctx)
								}

							} else if atomic.LoadInt32(&unstakesCount) < 4 && delegatedStakeAvailable > 0 && unclaimed < delegatedStakeAvailable/4 && delegatedUnstakePending == 0 && delegatedStakeAvailable > 5000 {
								if !testnet.mempool.ExistsTxSimpleVersion(addr.PublicKey, transaction_simple.SCRIPT_UNSTAKE) {
									if _, err = testnet.testnetCreateUnstakeTx(blockHeight, delegatedStakeAvailable/2-unclaimed); err != nil {
										return
									}
								}
								atomic.AddInt32(&unstakesCount, 1)
							} else {

								time.Sleep(time.Millisecond * 100) //making sure the block got propagated
								for i := 2; i < 5; i++ {
									testnet.testnetCreateTransfers(i, ctx)
								}

							}
						}

					}

				}

				return
			}(); err != nil {
				gui.GUI.Error("Error creating testnet Tx", err)
				err = nil
			}

		})

	}

	oldCancel()

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
