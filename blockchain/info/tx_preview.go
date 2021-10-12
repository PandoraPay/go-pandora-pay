package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/helpers"
)

type TxPreviewSimpleExtraUnstake struct {
	Amount uint64 `json:"amount"`
}

type TxPreviewSimpleOutput struct {
	PublicKey helpers.HexBytes `json:"publicKey"`
	Amount    uint64           `json:"amount"`
}

type TxPreviewSimpleExtraClaim struct {
	Output []*TxPreviewSimpleOutput `json:"output"`
}

type TxPreviewSimple struct {
	Extra      interface{}                   `json:"extra"`
	TxScript   transaction_simple.ScriptType `json:"txScript"`
	DataPublic helpers.HexBytes              `json:"dataPublic"`
	Fees       uint64                        `json:"fees"`
	Vin        helpers.HexBytes              `json:"vin"`
}

type TxPreviewZetherPayload struct {
	Asset      helpers.HexBytes   `json:"asset"`
	BurnValue  uint64             `json:"burnValue"`
	DataPublic helpers.HexBytes   `json:"dataPublic"`
	Publickeys []helpers.HexBytes `json:"publicKeys"`
	Fees       uint64             `json:"fees"`
}

type TxPreviewZether struct {
	TxScript transaction_zether.ScriptType `json:"txScript"`
	Payloads []*TxPreviewZetherPayload     `json:"payloads"`
}

type TxPreview struct {
	TxBase  interface{}                         `json:"base"`
	Version transaction_type.TransactionVersion `json:"version"`
	Hash    helpers.HexBytes                    `json:"hash"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base interface{}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var baseExtra interface{}
		switch txBase.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE: //nothing to be copied
		case transaction_simple.SCRIPT_CLAIM:
			txExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleClaim)
			extraClaim := &TxPreviewSimpleExtraClaim{
				Output: make([]*TxPreviewSimpleOutput, len(txExtra.Output)),
			}
			for i, out := range txExtra.Output {
				extraClaim.Output[i] = &TxPreviewSimpleOutput{out.PublicKey, out.Amount}
			}

			baseExtra = extraClaim
		case transaction_simple.SCRIPT_UNSTAKE:
			txExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleUnstake)
			baseExtra = &TxPreviewSimpleExtraUnstake{Amount: txExtra.Amount}
		}

		var dataPublic helpers.HexBytes
		if txBase.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataPublic = txBase.Data
		}

		base = &TxPreviewSimple{
			Extra:      baseExtra,
			TxScript:   txBase.TxScript,
			Fees:       txBase.Fees,
			Vin:        txBase.Vin.PublicKey,
			DataPublic: dataPublic,
		}

	case transaction_type.TX_ZETHER:
		txBase := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
		payloads := make([]*TxPreviewZetherPayload, len(txBase.Payloads))
		for i, payload := range txBase.Payloads {
			publicKeys := make([]helpers.HexBytes, len(payload.Statement.Publickeylist))
			for i, publicKeyPoint := range payload.Statement.Publickeylist {
				publicKeys[i] = publicKeyPoint.EncodeCompressed()
			}

			var dataPublic helpers.HexBytes
			if payload.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
				dataPublic = payload.Data
			}

			payloads[i] = &TxPreviewZetherPayload{
				payload.Asset,
				payload.BurnValue,
				dataPublic,
				publicKeys,
				payload.Statement.Fees,
			}
		}

		base = &TxPreviewZether{
			TxScript: txBase.TxScript,
			Payloads: payloads,
		}
	default:
		return nil, errors.New("Invalid tx.Version")
	}

	return &TxPreview{
		base,
		tx.Version,
		tx.Bloom.Hash,
	}, nil
}
