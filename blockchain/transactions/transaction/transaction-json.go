package transaction

import (
	"encoding/json"
	"errors"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_simple_parts "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

type json_Only_Transaction struct {
	Version transaction_type.TransactionVersion `json:"version"`
}

type json_Transaction struct {
	*json_Only_Transaction
	Size uint64           `json:"size"`
	Hash helpers.HexBytes `json:"hash"`
}

type json_Only_TransactionSimple struct {
	TxScript    transaction_simple.ScriptType           `json:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion"`
	Data        helpers.HexBytes                        `json:"data"`
	Nonce       uint64                                  `json:"nonce"`
	Fee         uint64                                  `json:"fee"`
	Vin         *json_TransactionSimpleInput            `json:"vin"`
}

type json_TransactionSimple struct {
	*json_Transaction
	*json_Only_TransactionSimple
}

type json_TransactionSimpleInput struct {
	PublicKey helpers.HexBytes `json:"publicKey,omitempty"` //32
	Signature helpers.HexBytes `json:"signature"`           //64
}

type json_Only_TransactionSimpleUpdateDelegate struct {
	NewPublicKey helpers.HexBytes `json:"newPublicKey"` //20 byte
	NewFee       uint16           `json:"newFee"`       //20 byte
}

type json_TransactionSimpleUpdateDelegate struct {
	*json_TransactionSimple
	*json_Only_TransactionSimpleUpdateDelegate
}

type json_Only_TransactionSimpleUnstake struct {
	Amount uint64 `json:"amount"`
}

type json_TransactionSimpleUnstake struct {
	*json_TransactionSimple
	*json_Only_TransactionSimpleUnstake
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	txJson := &json_Transaction{
		&json_Only_Transaction{
			tx.Version,
		},
		tx.Bloom.Size,
		tx.Bloom.Hash,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		vinJson := &json_TransactionSimpleInput{
			base.Vin.PublicKey,
			base.Vin.Signature,
		}

		simpleJson := &json_TransactionSimple{
			txJson,
			&json_Only_TransactionSimple{
				base.TxScript,
				base.DataVersion,
				base.Data,
				base.Nonce,
				base.Fee,
				vinJson,
			},
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUpdateDelegate)
			return json.Marshal(&json_TransactionSimpleUpdateDelegate{
				simpleJson,
				&json_Only_TransactionSimpleUpdateDelegate{
					extra.NewPublicKey,
					extra.NewFee,
				},
			})
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake)
			return json.Marshal(&json_TransactionSimpleUnstake{
				simpleJson,
				&json_Only_TransactionSimpleUnstake{
					extra.Amount,
				},
			})
		default:
			return nil, errors.New("Invalid base.TxScript")
		}

	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) UnmarshalJSON(data []byte) error {

	txOnlyJson := &json_Only_Transaction{}
	if err := json.Unmarshal(data, txOnlyJson); err != nil {
		return err
	}

	switch txOnlyJson.Version {
	case transaction_type.TX_SIMPLE:
	default:
		return errors.New("Invalid Version")
	}

	tx.Version = txOnlyJson.Version

	switch tx.Version {
	case transaction_type.TX_SIMPLE:

		simpleJson := &json_Only_TransactionSimple{}
		if err := json.Unmarshal(data, simpleJson); err != nil {
			return err
		}

		switch simpleJson.DataVersion {
		case transaction_data.TX_DATA_NONE:
			if simpleJson.Data != nil {
				return errors.New("tx.Data must be nil")
			}

		case transaction_data.TX_DATA_PLAIN_TEXT, transaction_data.TX_DATA_ENCRYPTED:
			if simpleJson.Data == nil || len(simpleJson.Data) == 0 || len(simpleJson.Data) > config.TRANSACTIONS_MAX_DATA_LENGTH {
				return errors.New("Invalid tx.Data length")
			}
		default:
			return errors.New("Invalid tx.DataVersion")
		}

		vin := &transaction_simple_parts.TransactionSimpleInput{
			PublicKey: simpleJson.Vin.PublicKey,
			Signature: simpleJson.Vin.Signature,
		}

		base := &transaction_simple.TransactionSimple{
			TxScript:    simpleJson.TxScript,
			DataVersion: simpleJson.DataVersion,
			Data:        simpleJson.Data,
			Nonce:       simpleJson.Nonce,
			Fee:         simpleJson.Fee,
			Vin:         vin,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:

			extraJson := &json_Only_TransactionSimpleUpdateDelegate{}
			if err := json.Unmarshal(data, extraJson); err != nil {
				return err
			}

			base.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUpdateDelegate{
				NewPublicKey: extraJson.NewPublicKey,
				NewFee:       extraJson.NewFee,
			}

		case transaction_simple.SCRIPT_UNSTAKE:
			extraJSON := &json_Only_TransactionSimpleUnstake{}
			if err := json.Unmarshal(data, extraJSON); err != nil {
				return err
			}

			base.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUnstake{
				Amount: extraJSON.Amount,
			}
		default:
			return errors.New("Invalid base.TxScript")
		}

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
