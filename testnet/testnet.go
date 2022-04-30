package testnet

import (
	"bytes"
	"context"
	"github.com/tevino/abool"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_builder/wizard"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
	"time"
)

type Testnet struct {
	wallet     *wallet.Wallet
	mempool    *mempool.Mempool
	chain      *blockchain.Blockchain
	txsBuilder *txs_builder.TxsBuilder
	nodes      uint64
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	//txData := &txs_builder.TxBuilderCreateZetherTxData{
	//	Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{},
	//}
	//
	//for i := uint64(0); i < testnet.nodes; i++ {
	//
	//	var addrSender, addrRecipient *wallet_address.WalletAddress
	//	if addrSender, err = testnet.wallet.GetWalletAddress(1, true); err != nil {
	//		return
	//	}
	//
	//	if addrRecipient, err = testnet.wallet.GetWalletAddress(int(i+1), true); err != nil {
	//		return
	//	}
	//
	//	txData.Payloads = append(txData.Payloads, &txs_builder.TxBuilderCreateZetherTxPayload{
	//		Sender:            addrSender.AddressEncoded,
	//		Asset:             config_coins.NATIVE_ASSET_FULL,
	//		Amount:            config_stake.GetRequiredStake(blockHeight),
	//		Recipient:         addrRecipient.AddressEncoded,
	//		RingConfiguration: testnet.testnetGetZetherRingConfiguration(),
	//	})
	//}
	//
	//if tx, err = testnet.txsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(string) {}); err != nil {
	//	return nil, err
	//}
	//
	//gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) testnetCreateUnstakeTx(senderAddr *wallet_address.WalletAddress, unstakeAmount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	txData := &txs_builder.TxBuilderCreateSimpleTx{
		0,
		&wizard.WizardTransactionData{nil, false},
		&wizard.WizardTransactionFee{0, 0, 0, true},
		&wizard.WizardTxSimpleExtraUnstake{
			nil,
			[]uint64{unstakeAmount},
		},
		[]*txs_builder.TxBuilderCreateSimpleTxVin{{
			senderAddr.AddressEncoded,
			0,
			config_coins.NATIVE_ASSET_FULL,
		}},
		[]*txs_builder.TxBuilderCreateSimpleTxVout{},
	}

	if tx, err = testnet.txsBuilder.CreateSimpleTx(txData, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Unstake Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, tx.Bloom.Hash)

	return
}

func (testnet *Testnet) testnetCreateTransfers(senderAddr *wallet_address.WalletAddress, amount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	var addr *addresses.Address
	vout := []*txs_builder.TxBuilderCreateSimpleTxVout{}
	total := uint64(0)

	for i := 0; i < 6; i++ {

		privateKey := addresses.GenerateNewPrivateKey()
		if addr, err = privateKey.GenerateAddress(nil, 0, nil); err != nil {
			return
		}

		if amount == 0 {
			amount = uint64(rand.Int63n(100))
		}
		total += amount
		vout = append(vout, &txs_builder.TxBuilderCreateSimpleTxVout{
			addr.EncodeAddr(),
			amount,
			config_coins.NATIVE_ASSET_FULL,
		})
	}

	txData := &txs_builder.TxBuilderCreateSimpleTx{
		0,
		&wizard.WizardTransactionData{nil, false},
		&wizard.WizardTransactionFee{0, 0, 0, true},
		nil,
		[]*txs_builder.TxBuilderCreateSimpleTxVin{{
			senderAddr.AddressEncoded,
			total,
			config_coins.NATIVE_ASSET_FULL,
		}},
		vout,
	}

	if tx, err = testnet.txsBuilder.CreateSimpleTx(txData, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) testnetCreateRecipientTransfers(senderAddr *wallet_address.WalletAddress, recipient int, amount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	var addrRecipient *wallet_address.WalletAddress
	if addrRecipient, err = testnet.wallet.GetWalletAddress(recipient, true); err != nil {
		return
	}

	txData := &txs_builder.TxBuilderCreateSimpleTx{
		0,
		&wizard.WizardTransactionData{nil, false},
		&wizard.WizardTransactionFee{0, 0, 0, true},
		nil,
		[]*txs_builder.TxBuilderCreateSimpleTxVin{{
			senderAddr.AddressEncoded,
			amount,
			config_coins.NATIVE_ASSET_FULL,
		}},
		[]*txs_builder.TxBuilderCreateSimpleTxVout{{
			addrRecipient.AddressEncoded,
			amount,
			config_coins.NATIVE_ASSET_FULL,
		}},
	}

	if tx, err = testnet.txsBuilder.CreateSimpleTx(txData, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) run() {

	updateChannel := testnet.chain.UpdateNewChainDataUpdate.AddListener()
	defer testnet.chain.UpdateNewChainDataUpdate.RemoveChannel(updateChannel)

	creatingTransactions := abool.New()

	for i := uint64(0); i < 10; i++ {
		if uint64(testnet.wallet.GetAddressesCount()) <= i+1 {
			if _, err := testnet.wallet.AddNewAddress(true, "Testnet wallet", true); err != nil {
				return
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {

		chainData, ok := <-updateChannel
		if !ok {
			return
		}

		syncTime := testnet.chain.Sync.GetSyncTime()

		blockHeight := chainData.Update.Height
		blockTimestamp := chainData.Update.Timestamp

		recovery.SafeGo(func() {

			gui.GUI.Log("UpdateNewChain received! 1")
			defer gui.GUI.Log("UpdateNewChain received! DONE")

			if err := func() (err error) {

				if blockHeight == 100 {
					if _, err = testnet.testnetCreateTransfersNewWallets(blockHeight, ctx); err != nil {
						return
					}
				}

				if blockTimestamp < uint64(time.Now().Unix()-10*60) {
					return
				}

				if blockHeight >= 20 {

					creatingTransactions.Set()
					defer creatingTransactions.UnSet()

					var addr, tempAddr *wallet_address.WalletAddress
					addr, _ = testnet.wallet.GetFirstStakedAddress(true)

					addressesList := []*wallet_address.WalletAddress{}
					for i := 0; i < 5; i++ {
						if tempAddr, err = testnet.wallet.GetWalletAddress(i, true); err != nil {
							return
						}
						addressesList = append(addressesList, tempAddr)
					}

					type AccMapElement struct {
						index          int
						balance        uint64
						stakeAvailable uint64
					}
					accMap := map[string]*AccMapElement{}

					gui.GUI.Log("UpdateNewChain received! 2")

					if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

						dataStorage := data_storage.NewDataStorage(reader)

						var accs *accounts.Accounts
						var acc *account.Account
						var plainAccount *plain_account.PlainAccount

						if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
							return
						}
						for i := 0; i < 5; i++ {

							if acc, err = accs.GetAccount(addressesList[i].PublicKeyHash); err != nil {
								return
							}
							accMap[string(addressesList[i].PublicKeyHash)] = &AccMapElement{
								i,
								0, 0,
							}

							if plainAccount, err = dataStorage.PlainAccs.GetPlainAccount(addressesList[i].PublicKeyHash); err != nil {
								return
							}
							if plainAccount != nil {
								accMap[string(addressesList[i].PublicKeyHash)].stakeAvailable = plainAccount.StakeAvailable
							}
							if acc != nil {
								accMap[string(addressesList[i].PublicKeyHash)].balance = acc.Balance
							}
						}

						return
					}); err != nil {
						return
					}

					stakingAmount := accMap[string(addr.PublicKeyHash)].stakeAvailable
					balance := accMap[string(addr.PublicKeyHash)].balance

					time.Sleep(time.Millisecond * 3000) //making sure the block got propagated

					if stakingAmount > config_coins.ConvertToUnitsUint64Forced(100000) {
						over := stakingAmount - config_coins.ConvertToUnitsUint64Forced(80000)
						testnet.testnetCreateUnstakeTx(addr, over, ctx)

						stakingAmount = generics.Max(0, stakingAmount-over)
					}

					if syncTime > 0 {

						if balance > config_coins.ConvertToUnitsUint64Forced(20000) {
							over := balance - config_coins.ConvertToUnitsUint64Forced(10000)
							if !testnet.mempool.ExistsTxSimpleVersion(addr.PublicKey, transaction_simple.SCRIPT_TRANSFER) {
								for i := 0; i < 5; i++ {

									if bytes.Equal(addr.PublicKey, addressesList[i].PublicKey) {
										continue
									}

									if accMap[string(addressesList[i].PublicKeyHash)].balance < config_coins.ConvertToUnitsUint64Forced(10000) {
										amount := generics.Min(over/5, config_coins.ConvertToUnitsUint64Forced(10000)-accMap[string(addressesList[i].PublicKeyHash)].balance)
										testnet.testnetCreateRecipientTransfers(addr, i, amount, ctx)
									}

								}
							}
						}

						for i := 1; i < 5; i++ {
							if bytes.Equal(addr.PublicKey, addressesList[i].PublicKey) {
								continue
							}
							for q := 0; q < 3; q++ {
								testnet.testnetCreateTransfers(addressesList[i], 0, ctx)
								time.Sleep(time.Millisecond * 200)
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

}

func TestnetInit(wallet *wallet.Wallet, mempool *mempool.Mempool, chain *blockchain.Blockchain, txsBuilder *txs_builder.TxsBuilder) (testnet *Testnet) {

	testnet = &Testnet{
		wallet:     wallet,
		mempool:    mempool,
		chain:      chain,
		txsBuilder: txsBuilder,
		nodes:      uint64(config.CPU_THREADS),
	}

	recovery.SafeGo(testnet.run)

	return
}
