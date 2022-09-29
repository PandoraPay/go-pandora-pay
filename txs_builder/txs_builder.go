package txs_builder

import (
	"context"
	"errors"
	"fmt"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_builder/wizard"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
	"sync"
)

type TxsBuilder struct {
	wallet       *wallet.Wallet
	txsValidator *txs_validator.TxsValidator
	mempool      *mempool.Mempool
	lock         *sync.Mutex
}

func (builder *TxsBuilder) getNonce(nonce uint64, publicKey []byte, accNonce uint64) uint64 {
	if nonce != 0 {
		return nonce
	}
	return builder.mempool.GetNonce(publicKey, accNonce)
}

func (builder *TxsBuilder) convertFloatAmounts(amounts []float64, ast *asset.Asset) ([]uint64, error) {

	var err error

	amountsFinal := make([]uint64, len(amounts))
	for i := range amounts {
		if err != nil {
			return nil, err
		}
		if amountsFinal[i], err = ast.ConvertToUnits(amounts[i]); err != nil {
			return nil, err
		}
	}

	return amountsFinal, nil
}

func (builder *TxsBuilder) getWalletAddresses(senders []string) ([]*wallet_address.WalletAddress, error) {

	sendersWalletAddress := make([]*wallet_address.WalletAddress, len(senders))
	var err error

	for i, senderAddress := range senders {
		if sendersWalletAddress[i], err = builder.wallet.GetWalletAddressByEncodedAddress(senderAddress, true); err != nil {
			return nil, err
		}
		if sendersWalletAddress[i].PrivateKey == nil {
			return nil, fmt.Errorf("Can't be used for transactions as the private key is missing for sender %s", senderAddress)
		}
	}

	return sendersWalletAddress, nil
}

func (builder *TxsBuilder) CreateSimpleTx(txData *TxBuilderCreateSimpleTx, propagateTx, awaitAnswer, awaitBroadcast, validateTx bool, ctx context.Context, statusCallback func(status string)) (*transaction.Transaction, error) {

	if txData.Data == nil {
		txData.Data = &wizard.WizardTransactionData{nil, false}
	}
	if txData.Fee == nil {
		txData.Fee = &wizard.WizardTransactionFee{0, 0, 0, true}
	}

	var sendersWalletAddresses []*wallet_address.WalletAddress
	var err error
	if txData.Sender != "" {
		if sendersWalletAddresses, err = builder.getWalletAddresses([]string{txData.Sender}); err != nil {
			return nil, err
		}
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	statusCallback("Wallet Addresses Found")

	transfer := &wizard.WizardTxSimpleTransfer{
		txData.Extra,
		txData.Data,
		txData.Fee,
		txData.Nonce,
		nil,
	}

	var tx *transaction.Transaction
	var plainAcc *plain_account.PlainAccount
	var chainHeight uint64

	if len(sendersWalletAddresses) > 0 {

		if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

			plainAccs := plain_accounts.NewPlainAccounts(reader)

			if plainAcc, err = plainAccs.Get(string(sendersWalletAddresses[0].PublicKey)); err != nil {
				return
			}
			if plainAcc == nil {
				return errors.New("Plain Account doesn't exist")
			}

			return
		}); err != nil {
			return nil, err
		}

		statusCallback("Getting Nonce from Mempool")
		transfer.Nonce = builder.getNonce(txData.Nonce, sendersWalletAddresses[0].PublicKey, plainAcc.Nonce)
		transfer.Key = sendersWalletAddresses[0].PrivateKey.Key
	}

	if tx, err = wizard.CreateSimpleTx(transfer, false, statusCallback); err != nil {
		return nil, err
	}
	statusCallback("Transaction Created")

	if propagateTx {
		if err = builder.mempool.AddTxToMempool(tx, chainHeight, true, awaitAnswer, awaitBroadcast, advanced_connection_types.UUID_ALL, ctx); err != nil {
			return nil, err
		}
	}

	return tx, nil
}

func TxsBuilderInit(wallet *wallet.Wallet, mempool *mempool.Mempool, txsValidator *txs_validator.TxsValidator) (builder *TxsBuilder) {

	builder = &TxsBuilder{
		wallet,
		txsValidator,
		mempool,
		&sync.Mutex{},
	}

	builder.initCLI()

	return
}
