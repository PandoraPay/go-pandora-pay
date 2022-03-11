package testnet

import (
	"context"
	"encoding/base64"
	"github.com/tevino/abool"
	"math"
	"math/rand"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_script"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
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

func (testnet *Testnet) testnetGetZetherRingConfiguration() *txs_builder.ZetherRingConfiguration {
	zetherRingConfiguration := &txs_builder.ZetherRingConfiguration{-1, &txs_builder.ZetherSenderRingType{}, &txs_builder.ZetherRecipientRingType{false, nil, -1}}
	if config.LIGHT_COMPUTATIONS {
		zetherRingConfiguration.RingSize = int(math.Pow(2, float64(rand.Intn(2)+3)))
	}
	return zetherRingConfiguration
}

func (testnet *Testnet) testnetCreateUnstakeTx(blockHeight uint64, amount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	addr, err := testnet.wallet.GetWalletAddress(0, true)
	if err != nil {
		return
	}

	if tx, err = testnet.txsBuilder.CreateSimpleTx(&txs_builder.TxBuilderCreateSimpleTx{addr.AddressEncoded, 0, nil, nil, false, &wizard.WizardTxSimpleExtraUnstake{Amount: amount}}, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Unstake tx was created: " + base64.StdEncoding.EncodeToString(tx.Bloom.Hash))
	return
}

func (testnet *Testnet) testnetCreateTransfersNewWallets(blockHeight uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	txData := &txs_builder.TxBuilderCreateZetherTxData{
		Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{},
	}

	for i := uint64(0); i < testnet.nodes; i++ {

		var addrSender, addrRecipient *wallet_address.WalletAddress
		if addrSender, err = testnet.wallet.GetWalletAddress(1, true); err != nil {
			return
		}

		if addrRecipient, err = testnet.wallet.GetWalletAddress(int(i+1), true); err != nil {
			return
		}

		txData.Payloads = append(txData.Payloads, &txs_builder.TxBuilderCreateZetherTxPayload{
			Sender:            addrSender.AddressEncoded,
			Asset:             config_coins.NATIVE_ASSET_FULL,
			Amount:            config_stake.GetRequiredStake(blockHeight),
			Recipient:         addrRecipient.AddressEncoded,
			RingConfiguration: testnet.testnetGetZetherRingConfiguration(),
		})
	}

	if tx, err = testnet.txsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) testnetCreateClaimTx(recipientAddressWalletIndex int, sendAmount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	var addrSender, addrRecipient *wallet_address.WalletAddress
	if addrSender, err = testnet.wallet.GetWalletAddress(0, true); err != nil {
		return
	}

	if addrRecipient, err = testnet.wallet.GetWalletAddress(recipientAddressWalletIndex, true); err != nil {
		return
	}

	var acc *account.Account
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		var accs *accounts.Accounts
		if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
			return
		}

		if acc, err = accs.GetAccount(addrRecipient.PublicKey); err != nil {
			return
		}
		return
	}); err != nil {
		return
	}

	var amount uint64
	if acc != nil {
		if amount, err = testnet.wallet.DecryptBalance(addrRecipient, acc.Balance.Amount.Serialize(), config_coins.NATIVE_ASSET_FULL, false, 0, true, ctx, func(string) {}); err != nil {
			return
		}
	}

	if amount > config_coins.ConvertToUnitsUint64Forced(10000) {
		return nil, nil
	}

	if sendAmount > amount+config_coins.ConvertToUnitsUint64Forced(10000) {
		sendAmount = (config_coins.ConvertToUnitsUint64Forced(10000) + amount) - sendAmount
	}

	txData := &txs_builder.TxBuilderCreateZetherTxData{
		Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{{
			Sender:            addrSender.AddressEncoded,
			Amount:            sendAmount,
			Recipient:         addrRecipient.AddressRegistrationEncoded,
			Data:              &wizard.WizardTransactionData{nil, false},
			Fee:               &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0},
			Asset:             config_coins.NATIVE_ASSET_FULL,
			RingConfiguration: testnet.testnetGetZetherRingConfiguration(),
		}},
	}

	if tx, err = testnet.txsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Claim Stake Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)

	return
}

func (testnet *Testnet) testnetCreateTransfers(senderAddressWalletIndex int, amount uint64, ctx context.Context) (tx *transaction.Transaction, err error) {

	select {
	case <-ctx.Done():
		return
	default:
	}

	senderAddr, err := testnet.wallet.GetWalletAddress(senderAddressWalletIndex, true)
	if err != nil {
		return
	}

	privateKey := addresses.GenerateNewPrivateKey()

	addr, err := privateKey.GenerateAddress(true, nil, 0, nil)
	if err != nil {
		return
	}

	if amount == 0 {
		amount = uint64(rand.Int63n(6))
	}

	txData := &txs_builder.TxBuilderCreateZetherTxData{
		Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{{
			Sender:            senderAddr.AddressEncoded,
			Amount:            amount,
			Recipient:         addr.EncodeAddr(),
			Data:              &wizard.WizardTransactionData{nil, false},
			Fee:               &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0},
			Asset:             config_coins.NATIVE_ASSET_FULL,
			RingConfiguration: testnet.testnetGetZetherRingConfiguration(),
		}},
	}

	if tx, err = testnet.txsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(string) {}); err != nil {
		return nil, err
	}

	gui.GUI.Info("Create Transfers Tx: ", tx.TransactionBaseInterface.(*transaction_zether.TransactionZether).ChainHeight, tx.Bloom.Hash)
	return
}

func (testnet *Testnet) run() {

	updateChannel := testnet.chain.UpdateNewChain.AddListener()
	defer testnet.chain.UpdateNewChain.RemoveChannel(updateChannel)

	creatingTransactions := abool.New()

	for i := uint64(0); i < testnet.nodes; i++ {
		if uint64(testnet.wallet.GetAddressesCount()) <= i+1 {
			if _, err := testnet.wallet.AddNewAddress(true, "Testnet wallet"); err != nil {
				return
			}
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for {

		blockHeight, ok := <-updateChannel
		if !ok {
			return
		}

		syncTime := testnet.chain.Sync.GetSyncTime()

		recovery.SafeGo(func() {

			gui.GUI.Log("UpdateNewChain received! 1")
			defer gui.GUI.Log("UpdateNewChain received! DONE")

			if err := func() (err error) {

				if blockHeight == 100 {
					if _, err = testnet.testnetCreateTransfersNewWallets(blockHeight, ctx); err != nil {
						return
					}
				}

				if blockHeight >= 30 {

					creatingTransactions.Set()
					defer creatingTransactions.UnSet()

					var addr *wallet_address.WalletAddress
					addr, _ = testnet.wallet.GetWalletAddress(0, true)

					var acc *account.Account

					gui.GUI.Log("UpdateNewChain received! 2")

					if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

						dataStorage := data_storage.NewDataStorage(reader)

						var accs *accounts.Accounts
						if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
							return
						}

						if acc, err = accs.GetAccount(addr.PublicKey); err != nil {
							return
						}
						return
					}); err != nil {
						return
					}

					var stakingAmount uint64
					if stakingAmount, err = testnet.wallet.DecryptBalance(addr, acc.Balance.Amount.Serialize(), config_coins.NATIVE_ASSET_FULL, false, 0, true, ctx, func(string) {}); err != nil {
						return
					}

					time.Sleep(time.Millisecond * 3000) //making sure the block got propagated

					if stakingAmount > config_coins.ConvertToUnitsUint64Forced(120000) {
						over := stakingAmount - config_coins.ConvertToUnitsUint64Forced(100000)
						testnet.testnetCreateTransfers(0, over, ctx)

						stakingAmount = generics.Max(0, config_coins.ConvertToUnitsUint64Forced(100000))
					}

					if syncTime > 0 {

						if stakingAmount > config_coins.ConvertToUnitsUint64Forced(20000) {
							over := stakingAmount - config_coins.ConvertToUnitsUint64Forced(10000)
							if !testnet.mempool.ExistsTxZetherVersion(addr.PublicKey, transaction_zether_payload_script.SCRIPT_TRANSFER) {
								testnet.testnetCreateClaimTx(1, over/5, ctx)
								testnet.testnetCreateClaimTx(2, over/5, ctx)
								testnet.testnetCreateClaimTx(3, over/5, ctx)
								testnet.testnetCreateClaimTx(4, over/5, ctx)
							}
						}

						for i := 2; i < 5; i++ {
							testnet.testnetCreateTransfers(i, 0, ctx)
							time.Sleep(time.Millisecond * 5000)
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
