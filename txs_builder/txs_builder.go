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

func (builder *TxsBuilder) getNonce(nonce uint64, publicKeyHash []byte, accNonce uint64) uint64 {
	if nonce != 0 {
		return nonce
	}
	return builder.mempool.GetNonce(publicKeyHash, accNonce)
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

	sendersWalletAddresses, err := builder.getWalletAddresses([]string{txData.Vin[0].Sender})
	if err != nil {
		return nil, err
	}

	builder.lock.Lock()
	defer builder.lock.Unlock()

	statusCallback("Wallet Addresses Found")

	var tx *transaction.Transaction
	var plainAcc *plain_account.PlainAccount
	var chainHeight uint64

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		plainAccs := plain_accounts.NewPlainAccounts(reader)

		for i := range sendersWalletAddresses {
			if plainAcc, err = plainAccs.GetPlainAccount(sendersWalletAddresses[i].PublicKeyHash); err != nil {
				return
			}
			if plainAcc == nil {
				return errors.New("Plain Account doesn't exist")
			}

			switch txExtra := txData.Extra.(type) {
			case *wizard.WizardTxSimpleExtraUnstake:
				if plainAcc.StakeAvailable < txExtra.Amounts[i] {
					return errors.New("You don't have enough staked coins")
				}
			}
		}

		return
	}); err != nil {
		return nil, err
	}

	statusCallback("Balances checked")

	txData.Nonce = builder.getNonce(txData.Nonce, sendersWalletAddresses[0].PublicKey, plainAcc.Nonce)
	statusCallback("Getting Nonce from Mempool")

	vin := make([]*wizard.WizardTxSimpleTransferVin, len(txData.Vin))
	for i, v := range txData.Vin {
		vin[i] = &wizard.WizardTxSimpleTransferVin{
			sendersWalletAddresses[i].PrivateKey.Key,
			v.Amount,
			v.Asset,
		}
	}

	vout := make([]*wizard.WizardTxSimpleTransferVout, len(txData.Vout))
	for i, v := range txData.Vout {
		vout[i] = &wizard.WizardTxSimpleTransferVout{
			v.PublicKeyHash,
			v.Amount,
			v.Asset,
		}
	}

	if tx, err = wizard.CreateSimpleTx(&wizard.WizardTxSimpleTransfer{
		txData.Extra,
		txData.Data,
		txData.Fee,
		txData.Nonce,
		vin,
		vout,
	}, false, statusCallback); err != nil {
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
