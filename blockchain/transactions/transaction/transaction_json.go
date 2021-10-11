package transaction

import (
	"encoding/json"
	"errors"
	"math"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type json_TransactionRegistration struct {
	PublicKeyIndex        uint64           `json:"publicKeyIndex"`
	RegistrationSignature helpers.HexBytes `json:"signature"`
}

type json_Transaction struct {
	Version       transaction_type.TransactionVersion `json:"version"`
	Registrations []*json_TransactionRegistration     `json:"registrations"`
	Size          uint64                              `json:"size"`
	Hash          helpers.HexBytes                    `json:"hash"`
}

type json_Only_TransactionSimple struct {
	TxScript    transaction_simple.ScriptType           `json:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion"`
	Data        helpers.HexBytes                        `json:"data"`
	Nonce       uint64                                  `json:"nonce"`
	Fees        uint64                                  `json:"fee"`
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
	NewFee       uint64           `json:"newFee"`       //20 byte
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

type json_Only_TransactionSimpleOutput struct {
	Amount    uint64           `json:"amount"`
	PublicKey helpers.HexBytes `json:"publicKey"`
}

type json_Only_TransactionSimpleClaim struct {
	Output []*json_Only_TransactionSimpleOutput `json:"output"`
}

type json_TransactionSimpleClaim struct {
	*json_TransactionSimple
	*json_Only_TransactionSimpleClaim
}

type json_Only_TransactionZether struct {
	TxScript transaction_zether.ScriptType   `json:"txScript"`
	Height   uint64                          `json:"height"`
	Payloads []*json_Only_TransactionPayload `json:"payloads"`
}

type json_Only_TransactionZetherStatement struct {
	RingSize      uint64             `json:"ringSize"`
	CLn           []helpers.HexBytes `json:"cLn"`
	CRn           []helpers.HexBytes `json:"cRn"`
	Publickeylist []helpers.HexBytes `json:"publickeylist"`
	C             []helpers.HexBytes `json:"c"`
	D             helpers.HexBytes   `json:"d"`
	Fees          uint64             `json:"fees"`
}

type json_Only_TransactionPayload struct {
	Token       helpers.HexBytes                        `json:"token"`
	BurnValue   uint64                                  `json:"burnValue"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataType"`
	Data        helpers.HexBytes                        `json:"data"`
	Statement   *json_Only_TransactionZetherStatement   `json:"statement"`
	Proof       helpers.HexBytes                        `json:"proof"`
}

type json_TransactionZether struct {
	*json_Transaction
	*json_Only_TransactionZether
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	registrations := make([]*json_TransactionRegistration, len(tx.Registrations.Registrations))
	for i, reg := range tx.Registrations.Registrations {
		registrations[i] = &json_TransactionRegistration{
			reg.PublicKeyIndex,
			reg.RegistrationSignature,
		}
	}

	txJson := &json_Transaction{
		tx.Version,
		registrations,
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
				base.Fees,
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
		case transaction_simple.SCRIPT_CLAIM:
			extra := base.TransactionSimpleExtraInterface.(*transaction_simple_extra.TransactionSimpleClaim)

			output := make([]*json_Only_TransactionSimpleOutput, len(extra.Output))
			for i, out := range extra.Output {
				output[i] = &json_Only_TransactionSimpleOutput{
					out.Amount,
					out.PublicKey,
				}
			}

			return json.Marshal(&json_TransactionSimpleClaim{
				simpleJson,
				&json_Only_TransactionSimpleClaim{
					output,
				},
			})
		default:
			return nil, errors.New("Invalid base.TxScript")
		}
	case transaction_type.TX_ZETHER:
		base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

		payloadsJson := make([]*json_Only_TransactionPayload, len(base.Payloads))
		for i, payload := range base.Payloads {

			statementJson := &json_Only_TransactionZetherStatement{
				RingSize:      payload.Statement.RingSize,
				CLn:           helpers.ConvertBN256Array(payload.Statement.CLn),
				CRn:           helpers.ConvertBN256Array(payload.Statement.CRn),
				Publickeylist: helpers.ConvertBN256Array(payload.Statement.Publickeylist),
				C:             helpers.ConvertBN256Array(payload.Statement.C),
				D:             payload.Statement.D.EncodeCompressed(),
				Fees:          payload.Statement.Fees,
			}

			w := helpers.NewBufferWriter()
			payload.Proof.Serialize(w)
			proofJson := w.Bytes()

			payloadsJson[i] = &json_Only_TransactionPayload{
				payload.Token,
				payload.BurnValue,
				payload.DataVersion,
				payload.Data,
				statementJson,
				proofJson,
			}
		}

		zetherJson := &json_TransactionZether{
			txJson,
			&json_Only_TransactionZether{
				base.TxScript,
				base.Height,
				payloadsJson,
			},
		}
		return json.Marshal(zetherJson)
	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) UnmarshalJSON(data []byte) error {

	txOnlyJson := &json_Transaction{}
	if err := json.Unmarshal(data, txOnlyJson); err != nil {
		return err
	}

	switch txOnlyJson.Version {
	case transaction_type.TX_SIMPLE, transaction_type.TX_ZETHER:
	default:
		return errors.New("Invalid Version")
	}

	tx.Version = txOnlyJson.Version

	tx.Registrations = &transaction_data.TransactionDataTransactions{
		Registrations: make([]*transaction_data.TransactionDataRegistration, len(txOnlyJson.Registrations)),
	}
	for i, reg := range txOnlyJson.Registrations {
		tx.Registrations.Registrations[i] = &transaction_data.TransactionDataRegistration{
			reg.PublicKeyIndex,
			reg.RegistrationSignature,
		}
	}

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
			Fees:        simpleJson.Fees,
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
		case transaction_simple.SCRIPT_CLAIM:
			extraJSON := &json_Only_TransactionSimpleClaim{}
			if err := json.Unmarshal(data, extraJSON); err != nil {
				return err
			}

			output := make([]*transaction_simple_parts.TransactionSimpleOutput, len(extraJSON.Output))
			for i, out := range extraJSON.Output {
				output[i] = &transaction_simple_parts.TransactionSimpleOutput{
					out.Amount,
					out.PublicKey,
				}
			}

			base.TransactionSimpleExtraInterface = &transaction_simple_extra.TransactionSimpleClaim{
				Output: output,
			}

		default:
			return errors.New("Invalid json Simple TxScript")
		}

	case transaction_type.TX_ZETHER:

		simpleZether := &json_Only_TransactionZether{}
		if err := json.Unmarshal(data, simpleZether); err != nil {
			return err
		}

		payloads := make([]*transaction_zether.TransactionZetherPayload, len(simpleZether.Payloads))
		for i, payload := range simpleZether.Payloads {

			CLn, err := helpers.ConvertToBN256Array(payload.Statement.CLn)
			if err != nil {
				return err
			}
			CRn, err := helpers.ConvertToBN256Array(payload.Statement.CRn)
			if err != nil {
				return err
			}
			Publickeylist, err := helpers.ConvertToBN256Array(payload.Statement.Publickeylist)
			if err != nil {
				return err
			}
			C, err := helpers.ConvertToBN256Array(payload.Statement.C)
			if err != nil {
				return err
			}

			D := new(bn256.G1)
			if err = D.DecodeCompressed(payload.Statement.D); err != nil {
				return err
			}

			m := int(math.Log2(float64(payload.Statement.RingSize)))
			if math.Pow(2, float64(m)) != float64(payload.Statement.RingSize) {
				return errors.New("log failed")
			}

			proof := &crypto.Proof{}
			if err = proof.Deserialize(helpers.NewBufferReader(payload.Proof), m); err != nil {
				return err
			}

			payloads[i] = &transaction_zether.TransactionZetherPayload{
				Token:       payload.Token,
				BurnValue:   payload.BurnValue,
				DataVersion: payload.DataVersion,
				Data:        payload.Data,
				Statement: &crypto.Statement{
					RingSize:      payload.Statement.RingSize,
					CLn:           CLn,
					CRn:           CRn,
					Publickeylist: Publickeylist,
					C:             C,
					D:             D,
					Fees:          payload.Statement.Fees,
				},
				Proof: proof,
			}
		}

		base := &transaction_zether.TransactionZether{
			TxScript: simpleZether.TxScript,
			Height:   simpleZether.Height,
			Payloads: payloads,
		}
		tx.TransactionBaseInterface = base

		switch simpleZether.TxScript {
		case transaction_zether.SCRIPT_TRANSFER:
		case transaction_zether.SCRIPT_DELEGATE:
		default:
			return errors.New("Invalid Zether TxScript")
		}

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
