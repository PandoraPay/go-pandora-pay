package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
	"pandora-pay/helpers"
)

type TxPreviewBase interface{}
type TxPreviewSimpleExtra interface{}

type TxPreviewSimpleExtraUnstake struct {
	TxPreviewSimpleExtra
	Amount uint64 `json:"amount"`
}

type TxPreviewSimpleExtraClaim struct {
	TxPreviewSimpleExtra
	PublicKeys []helpers.HexBytes `json:"publicKeys"`
	Amounts    []uint64           `json:"amounts"`
}

type TxPreviewSimple struct {
	TxPreviewBase
	Extra      TxPreviewSimpleExtra
	TxScript   transaction_simple.ScriptType `json:"txScript"`
	DataPublic helpers.HexBytes              `json:"dataPublic"`
	Fees       uint64                        `json:"fees"`
	Vin        helpers.HexBytes              `json:"vin"`
}

type TxPreviewZetherPayload struct {
	Token      helpers.HexBytes   `json:"token"`
	BurnValue  uint64             `json:"burnValue"`
	DataPublic helpers.HexBytes   `json:"dataPublic"`
	Publickeys []helpers.HexBytes `json:"publicKeys"`
	Fees       uint64             `json:"fees"`
}

type TxPreviewZether struct {
	TxPreviewBase
	TxScript transaction_zether.ScriptType `json:"txScript"`
	Payloads []*TxPreviewZetherPayload     `json:"payloads"`
}

type TxPreview struct {
	TxPreviewBase
	Version transaction_type.TransactionVersion `json:"version"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base TxPreviewBase
	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var baseExtra TxPreviewSimpleExtra
		switch txBase.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE: //nothing to be copied
		case transaction_simple.SCRIPT_CLAIM:
			txExtra := txBase.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleClaim)
			baseExtra = &TxPreviewSimpleExtraClaim{
				PublicKeys: make([]helpers.HexBytes, len(txExtra.Output)),
				Amounts:    make([]uint64, len(txExtra.Output)),
			}
		case transaction_simple.SCRIPT_UNSTAKE:
			txExtra := txBase.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake)
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
				payload.Token,
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
	}, nil
}
