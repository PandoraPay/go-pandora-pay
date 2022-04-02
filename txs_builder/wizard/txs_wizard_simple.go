package wizard

import (
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

	var txScript transaction_simple.ScriptType
	var extraFinal transaction_simple_extra.TransactionSimpleExtraInterface

	switch txExtra := transfer.Extra.(type) {
	case *WizardTxSimpleExtraUpdateAssetFeeLiquidity:
		extraFinal = &transaction_simple_extra.TransactionSimpleExtraUpdateAssetFeeLiquidity{
			Liquidities:     txExtra.Liquidities,
			CollectorHasNew: txExtra.CollectorHasNew,
			Collector:       txExtra.Collector,
		}
		txScript = transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY

		spaceExtra += 1 + len(txExtra.Collector) + 1
		for _, liquidity := range txExtra.Liquidities {
			if liquidity.Rate > 0 {
				spaceExtra += len(helpers.SerializeToBytes(liquidity))
			}
		}
	}

	txBase := &transaction_simple.TransactionSimple{
		TxScript:    txScript,
		DataVersion: transfer.Data.getDataVersion(),
		Data:        dataFinal,
		Nonce:       transfer.Nonce,
		Fee:         0,
		Extra:       extraFinal,
		Vin: &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: privateKey.GeneratePublicKey(),
		},
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
	var privateKey *addresses.PrivateKey
	if privateKey, err = addresses.NewPrivateKey(transfer.Key); err != nil {
		return nil, err
	}
	if txBase.Vin.Signature, err = privateKey.Sign(tx.SerializeForSigning()); err != nil {
		return nil, err
	}
	statusCallback("Transaction Signed")

	if err = bloomAllTx(tx, statusCallback); err != nil {
		return
	}
	return tx, nil
}
