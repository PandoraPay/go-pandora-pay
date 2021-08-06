package transaction

import (
	"encoding/json"
	"errors"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
)

type transactionJSON struct {
	Version transaction_type.TransactionVersion `json:"version"`
	Size    uint64                              `json:"size"`
	Hash    helpers.HexBytes                    `json:"hash"`
}

type transactionSimpleJSON struct {
	*transactionJSON
	TxScript transaction_simple.ScriptType  `json:"txScript"`
	Nonce    uint64                         `json:"nonce"`
	Vin      []*transactionSimpleInputJSON  `json:"vin"`
	Vout     []*transactionSimpleOutputJSON `json:"vout"`
}

type transactionSimpleInputJSON struct {
	Amount        uint64           `json:"amount"`
	Token         helpers.HexBytes `json:"token"`                   //20
	Signature     helpers.HexBytes `json:"signature"`               //65
	PublicKey     helpers.HexBytes `json:"publicKey,omitempty"`     //32
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash,omitempty"` //20
}

type transactionSimpleOutputJSON struct {
	PublicKeyHash helpers.HexBytes `json:"publicKeyHash"` //20
	Amount        uint64           `json:"amount"`
	Token         helpers.HexBytes `json:"token"` //20
}

type transactionSimpleDelegateJSON struct {
	*transactionSimpleJSON
	Amount              uint64           `json:"amount"`
	HasNewPublicKeyHash bool             `json:"hasNewPublicKeyHash"`
	NewPublicKeyHash    helpers.HexBytes `json:"newPublicKeyHash"` //20 byte
}

type transactionSimpleUnstakeJSON struct {
	*transactionSimpleJSON
	Amount   uint64 `json:"amount"`
	FeeExtra uint64 `json:"feeExtra"` //this will be subtracted StakeAvailable
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	data := &transactionJSON{
		Version: tx.Version,
		Size:    tx.Bloom.Size,
		Hash:    tx.Bloom.Hash,
	}

	switch tx.Version {
	case transaction_type.TX_SIMPLE:
		base := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)

		vinJSON := make([]*transactionSimpleInputJSON, len(base.Vin))
		for i, it := range base.Vin {
			vinJSON[i] = &transactionSimpleInputJSON{
				it.Amount,
				it.Token,
				it.Signature,
				it.Bloom.PublicKey,
				it.Bloom.PublicKeyHash,
			}
		}

		voutJSON := make([]*transactionSimpleOutputJSON, len(base.Vout))
		for i, it := range base.Vout {
			voutJSON[i] = &transactionSimpleOutputJSON{
				it.PublicKeyHash,
				it.Amount,
				it.Token,
			}
		}

		simpleJSON := &transactionSimpleJSON{
			data,
			base.TxScript,
			base.Nonce,
			vinJSON,
			voutJSON,
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_NORMAL:
			return json.Marshal(simpleJSON)
		case transaction_simple.SCRIPT_DELEGATE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleDelegate)
			return json.Marshal(&transactionSimpleDelegateJSON{
				simpleJSON,
				extra.Amount,
				extra.HasNewPublicKeyHash,
				extra.NewPublicKeyHash,
			})
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleUnstake)
			return json.Marshal(&transactionSimpleUnstakeJSON{
				simpleJSON,
				extra.Amount,
				extra.FeeExtra,
			})
		case transaction_simple.SCRIPT_WITHDRAW:
			return json.Marshal(simpleJSON)
		default:
			return nil, errors.New("Invalid base.TxScript")
		}

	default:
		return nil, errors.New("Invalid Tx Version")
	}

}
