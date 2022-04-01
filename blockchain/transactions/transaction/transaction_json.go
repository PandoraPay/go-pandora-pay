package transaction

import (
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/config"
)

type Json_Transaction struct {
	Version    transaction_type.TransactionVersion `json:"version" msgpack:"version"`
	Size       uint64                              `json:"size" msgpack:"size"`
	SpaceExtra uint64                              `json:"spaceExtra" msgpack:"spaceExtra"`
	Hash       []byte                              `json:"hash" msgpack:"hash"`
}

type json_TransactionSimple struct {
	*Json_Transaction
	TxScript    transaction_simple.ScriptType           `json:"txScript" msgpack:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion" msgpack:"dataVersion"`
	Data        []byte                                  `json:"data" msgpack:"data"`
	Nonce       uint64                                  `json:"nonce" msgpack:"nonce"`
	Fee         uint64                                  `json:"fee" msgpack:"fee"`
	Vin         *json_TransactionSimpleInput            `json:"vin" msgpack:"vin"`
	Extra       interface{}                             `json:"extra" msgpack:"extra"`
}

type json_TransactionSimpleInput struct {
	PublicKey []byte `json:"publicKey,omitempty" msgpack:"publicKey,omitempty"` //32
	Signature []byte `json:"signature" msgpack:"signature"`                     //64
}

func marshalJSON(tx *Transaction, marshal func(any) ([]byte, error)) ([]byte, error) {

	txJson := &Json_Transaction{
		tx.Version,
		tx.Bloom.Size,
		tx.SpaceExtra,
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
			base.TxScript,
			base.DataVersion,
			base.Data,
			base.Nonce,
			base.Fee,
			vinJson,
			nil,
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_TRANSFER:
		default:
			return nil, errors.New("Invalid simple.TxScript")
		}

		return marshal(simpleJson)

	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) MarshalJSON() ([]byte, error) {
	return marshalJSON(tx, json.Marshal)
}

func (tx *Transaction) EncodeMsgpack(enc *msgpack.Encoder) error {
	bytes, err := marshalJSON(tx, json.Marshal)
	if err != nil {
		return err
	}

	return enc.EncodeBytes(bytes)
}

func (tx *Transaction) UnmarshalJSON(data []byte) (err error) {

	txOnlyJson := &Json_Transaction{}
	if err = json.Unmarshal(data, txOnlyJson); err != nil {
		return
	}

	switch txOnlyJson.Version {
	case transaction_type.TX_SIMPLE:
	default:
		return errors.New("Invalid Version")
	}

	tx.Version = txOnlyJson.Version
	tx.SpaceExtra = txOnlyJson.SpaceExtra

	switch tx.Version {
	case transaction_type.TX_SIMPLE:

		simpleJson := &json_TransactionSimple{}
		if err = json.Unmarshal(data, simpleJson); err != nil {
			return
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
			nil,
			nil,
			simpleJson.TxScript,
			simpleJson.DataVersion,
			simpleJson.Data,
			simpleJson.Nonce,
			simpleJson.Fee,
			vin,
			nil,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_TRANSFER:
		default:
			return errors.New("Invalid json Simple TxScript")
		}

	default:
		return errors.New("Invalid Version")
	}

	return nil
}

func (tx *Transaction) DecodeMsgpack(dec *msgpack.Decoder) error {
	bytes, err := dec.DecodeBytes()
	if err != nil {
		return err
	}

	return tx.UnmarshalJSON(bytes)
}
