package transaction

import (
	"encoding/json"
	"errors"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_simple_parts "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
)

type json_TransactionVersion struct {
	Version transaction_type.TransactionVersion `json:"version"`
}

type json_Transaction struct {
	*json_TransactionVersion
	Size uint64           `json:"size"`
	Hash helpers.HexBytes `json:"hash"`
}

type json_Only_TransactionSimple struct {
	TxScript transaction_simple.ScriptType   `json:"txScript"`
	Nonce    uint64                          `json:"nonce"`
	Vin      []*json_TransactionSimpleInput  `json:"vin"`
	Vout     []*json_TransactionSimpleOutput `json:"vout"`
}

type json_TransactionSimple struct {
	*json_Transaction
	*json_Only_TransactionSimple
}

type json_TransactionSimpleInput struct {
	Amount        uint64           `json:"amount"`
	Token         helpers.HexBytes `json:"token"`                   //20
	Signature     helpers.HexBytes `json:"signature"`               //65
	PublicKey     helpers.HexBytes `json:"publicKey,omitempty"`     //32
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash,omitempty"` //20
}

type json_TransactionSimpleOutput struct {
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash"` //20
	Amount        uint64           `json:"amount"`
	Token         helpers.HexBytes `json:"token"` //20
}

type json_Only_TransactionSimpleDelegate struct {
	Amount              uint64           `json:"amount"`
	HasNewPublicKeyHash bool             `json:"hasNewPublicKeyHash"`
	NewPublicKeyHash    helpers.HexBytes `json:"newPublicKeyHash"` //20 byte
}

type json_TransactionSimpleDelegate struct {
	*json_TransactionSimple
	*json_Only_TransactionSimpleDelegate
}

type json_Only_TransactionSimpleUnstake struct {
	Amount   uint64 `json:"amount"`
	FeeExtra uint64 `json:"feeExtra"` //this will be subtracted StakeAvailable
}

type json_TransactionSimpleUnstake struct {
	*json_TransactionSimple
	*json_Only_TransactionSimpleUnstake
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	txJson := &json_Transaction{
		&json_TransactionVersion{
			Version: tx.Version,
		},
		tx.Bloom.Size,
		tx.Bloom.Hash,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		vinJson := make([]*json_TransactionSimpleInput, len(base.Vin))
		for i, it := range base.Vin {
			vinJson[i] = &json_TransactionSimpleInput{
				it.Amount,
				it.Token,
				it.Signature,
				it.Bloom.PublicKey,
				it.Bloom.PublicKeyHash,
			}
		}

		voutJson := make([]*json_TransactionSimpleOutput, len(base.Vout))
		for i, it := range base.Vout {
			voutJson[i] = &json_TransactionSimpleOutput{
				it.PublicKeyHash,
				it.Amount,
				it.Token,
			}
		}

		simpleJson := &json_TransactionSimple{
			txJson,
			&json_Only_TransactionSimple{
				base.TxScript,
				base.Nonce,
				vinJson,
				voutJson,
			},
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_NORMAL:
			return json.Marshal(simpleJson)
		case transaction_simple.SCRIPT_DELEGATE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleDelegate)
			return json.Marshal(&json_TransactionSimpleDelegate{
				simpleJson,
				&json_Only_TransactionSimpleDelegate{
					extra.Amount,
					extra.HasNewPublicKeyHash,
					extra.NewPublicKeyHash,
				},
			})
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake)
			return json.Marshal(&json_TransactionSimpleUnstake{
				simpleJson,
				&json_Only_TransactionSimpleUnstake{
					extra.Amount,
					extra.FeeExtra,
				},
			})
		case transaction_simple.SCRIPT_WITHDRAW:
			return json.Marshal(simpleJson)
		default:
			return nil, errors.New("Invalid base.TxScript")
		}

	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) UnmarshalJSON(data []byte) error {

	txVersionJson := &json_TransactionVersion{}
	if err := json.Unmarshal(data, txVersionJson); err != nil {
		return err
	}

	tx.Version = txVersionJson.Version
	switch tx.Version {
	case transaction_type.TX_SIMPLE:

		simpleJson := &json_Only_TransactionSimple{}
		if err := json.Unmarshal(data, simpleJson); err != nil {
			return err
		}

		vin := make([]*transaction_simple_parts.TransactionSimpleInput, len(simpleJson.Vin))
		for i, it := range simpleJson.Vin {
			vin[i] = &transaction_simple_parts.TransactionSimpleInput{
				Amount:    it.Amount,
				Token:     it.Token,
				Signature: it.Signature,
			}
		}

		vout := make([]*transaction_simple_parts.TransactionSimpleOutput, len(simpleJson.Vout))
		for i, it := range simpleJson.Vin {
			vout[i] = &transaction_simple_parts.TransactionSimpleOutput{
				PublicKeyHash: it.PublicKeyHash,
				Amount:        it.Amount,
				Token:         it.Token,
			}
		}

		base := &transaction_simple.TransactionSimple{
			TxScript: simpleJson.TxScript,
			Nonce:    simpleJson.Nonce,
			Vin:      vin,
			Vout:     vout,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_NORMAL:
		case transaction_simple.SCRIPT_DELEGATE:

			extraJson := &json_Only_TransactionSimpleDelegate{}
			if err := json.Unmarshal(data, extraJson); err != nil {
				return err
			}

			base.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleDelegate{
				Amount:              extraJson.Amount,
				HasNewPublicKeyHash: extraJson.HasNewPublicKeyHash,
				NewPublicKeyHash:    extraJson.NewPublicKeyHash,
			}

		case transaction_simple.SCRIPT_UNSTAKE:
			extraJSON := &json_Only_TransactionSimpleUnstake{}
			if err := json.Unmarshal(data, extraJSON); err != nil {
				return err
			}

			base.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleUnstake{
				Amount:   extraJSON.Amount,
				FeeExtra: extraJSON.FeeExtra,
			}
		case transaction_simple.SCRIPT_WITHDRAW:
		default:
			return errors.New("Invalid base.TxScript")
		}

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
