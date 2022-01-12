package info

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/helpers"
)

type TxPreviewSimpleExtraUnstake struct {
	Amount uint64 `json:"amount" msgpack:"amount"`
}

type TxPreviewSimple struct {
	Extra      interface{}                   `json:"extra" msgpack:"extra"`
	TxScript   transaction_simple.ScriptType `json:"txScript" msgpack:"txScript"`
	DataPublic helpers.HexBytes              `json:"dataPublic" msgpack:"dataPublic"`
	Vin        helpers.HexBytes              `json:"vin" msgpack:"vin"`
}

type TxPreviewZetherPayload struct {
	PayloadScript transaction_zether_payload.PayloadScriptType `json:"payloadScript" msgpack:"payloadScript"`
	Asset         helpers.HexBytes                             `json:"asset" msgpack:"asset"`
	BurnValue     uint64                                       `json:"burnValue" msgpack:"burnValue"`
	DataPublic    helpers.HexBytes                             `json:"dataPublic" msgpack:"dataPublic"`
	Publickeys    []helpers.HexBytes                           `json:"publicKeys" msgpack:"publicKeys"`
}

type TxPreviewZether struct {
	Payloads []*TxPreviewZetherPayload `json:"payloads"  msgpack:"payloads"`
}

type TxPreview struct {
	TxBase  interface{}                         `json:"base"  msgpack:"base"`
	Version transaction_type.TransactionVersion `json:"version"  msgpack:"version"`
	Hash    helpers.HexBytes                    `json:"hash"  msgpack:"hash"`
	Fee     uint64                              `json:"fee"  msgpack:"fee"`
}

func CreateTxPreviewFromTx(tx *transaction.Transaction) (*TxPreview, error) {

	var base interface{}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		var baseExtra interface{}
		switch txBase.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE: //nothing to be copied

		case transaction_simple.SCRIPT_UNSTAKE:
			txExtra := txBase.Extra.(*transaction_simple_extra.TransactionSimpleExtraUnstake)
			baseExtra = &TxPreviewSimpleExtraUnstake{Amount: txExtra.Amount}
		}

		var dataPublic helpers.HexBytes
		if txBase.DataVersion == transaction_data.TX_DATA_PLAIN_TEXT {
			dataPublic = txBase.Data
		}

		base = &TxPreviewSimple{
			Extra:      baseExtra,
			TxScript:   txBase.TxScript,
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
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				dataPublic,
				publicKeys,
			}

		}

		base = &TxPreviewZether{
			Payloads: payloads,
		}
	default:
		return nil, errors.New("Invalid tx.Version")
	}

	fee, err := tx.ComputeFee()
	if err != nil {
		return nil, err
	}

	return &TxPreview{
		base,
		tx.Version,
		tx.Bloom.Hash,
		fee,
	}, nil
}
