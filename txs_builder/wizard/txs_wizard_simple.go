package wizard

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

func CreateSimpleTx(transfer *WizardTxSimpleTransfer, validateTx bool, statusCallback func(string)) (tx2 *transaction.Transaction, err error) {

	dataFinal, err := transfer.Data.getData()
	if err != nil {
		return
	}

	spaceExtra := 0

	txBase := &transaction_simple.TransactionSimple{
		nil,
		nil,
		0,
		transfer.Data.getDataVersion(),
		dataFinal,
		transfer.Nonce,
		0,
		nil, nil,
	}

	switch txExtra := transfer.Extra.(type) {
	case *WizardTxSimpleExtraUpdateAssetFeeLiquidity:
		txBase.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{nil,
			txExtra.Liquidities,
			txExtra.NewCollector,
			txExtra.Collector,
		}
		txBase.TxScript = transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY

		spaceExtra += 1 + len(txExtra.Collector) + 1
		for _, liquidity := range txExtra.Liquidities {
			if liquidity.Rate > 0 {
				spaceExtra += len(helpers.SerializeToBytes(liquidity))
			}
		}
	case *WizardTxSimpleExtraResolutionPayInFuture:
		txBase.Extra = &transaction_simple_extra.TransactionSimpleExtraResolutionPayInFuture{nil,
			txExtra.TxId,
			txExtra.PayloadIndex,
			txExtra.Resolution,
			txExtra.MultisigPublicKeys,
			txExtra.Nonces,
			txExtra.Signatures,
		}
		txBase.TxScript = transaction_simple.SCRIPT_RESOLUTION_PAY_IN_FUTURE
	}

	var privateKey *addresses.PrivateKey

	switch txBase.TxScript {
	case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
		if privateKey, err = addresses.NewPrivateKey(transfer.Key); err != nil {
			return nil, err
		}

		txBase.Vin = &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		}

	case transaction_simple.SCRIPT_RESOLUTION_PAY_IN_FUTURE:
	default:
		return nil, errors.New("Invalid Tx Script")
	}

	tx := &transaction.Transaction{
		Version:                  transaction_type.TX_SIMPLE,
		SpaceExtra:               uint64(spaceExtra),
		TransactionBaseInterface: txBase,
	}
	statusCallback("Transaction Created")

	extraBytes := cryptography.SignatureSize
	txBase.Fee = setFee(tx, extraBytes, transfer.Fee.Clone(), true)
	statusCallback("Transaction Fee set")

	statusCallback("Transaction Signing...")

	if privateKey != nil {
		if txBase.Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
			return nil, err
		}
		statusCallback("Transaction Signed")
	}

	if err = bloomAllTx(tx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
