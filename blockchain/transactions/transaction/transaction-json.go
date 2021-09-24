package transaction

import (
	"encoding/json"
	"errors"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_simple_extra "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-extra"
	transaction_simple_parts "pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
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

type json_Only_TransactionPayloadStatement struct {
	RingSize      uint64             `json:"ringSize"`
	CLn           []helpers.HexBytes `json:"CLn"`
	CRn           []helpers.HexBytes `json:"CRn"`
	Publickeylist []helpers.HexBytes `json:"publickeylist"`
	C             []helpers.HexBytes `json:"C"`
	D             helpers.HexBytes   `json:"D"`
	Fees          uint64             `json:"fees"`
	Roothash      helpers.HexBytes   `json:"roothash"`
}

type InnerProduct struct {
	A  []helpers.HexBytes `json:"a"`
	B  []helpers.HexBytes `json:"b"`
	Ls []helpers.HexBytes `json:"ls"`
	Rs []helpers.HexBytes `json:"rs"`
}

type json_Only_TransactionPayloadProof struct {
	BA    helpers.HexBytes   `json:"BA"`
	BS    helpers.HexBytes   `json:"BS"`
	A     helpers.HexBytes   `json:"A"`
	B     helpers.HexBytes   `json:"B"`
	CLnG  []helpers.HexBytes `json:"CLnG"`
	CRnG  []helpers.HexBytes `json:"CRnG"`
	C_0G  []helpers.HexBytes `json:"C_0G"`
	DG    []helpers.HexBytes `json:"D_0G"`
	Y_0G  []helpers.HexBytes `json:"y_0G"`
	GG    []helpers.HexBytes `json:"gG"`
	C_XG  []helpers.HexBytes `json:"C_XG"`
	Y_XG  []helpers.HexBytes `json:"y_XG"`
	U     helpers.HexBytes   `json:"u"`
	U1    helpers.HexBytes   `json:"u1"`
	F     []helpers.HexBytes `json:"f"`
	Z_A   helpers.HexBytes   `json:"z_A"`
	T_1   helpers.HexBytes   `json:"T_1"`
	T_2   helpers.HexBytes   `json:"T_2"`
	That  helpers.HexBytes   `json:"that"`
	Mu    helpers.HexBytes   `json:"mu"`
	C     helpers.HexBytes   `json:"c"`
	S_sk  helpers.HexBytes   `json:"s_sk"`
	S_r   helpers.HexBytes   `json:"s_r"`
	S_b   helpers.HexBytes   `json:"s_b"`
	S_tau helpers.HexBytes   `json:"s_tau"`
	Ip    *InnerProduct      `json:"ip"`
}

type json_Only_TransactionPayload struct {
	Token     []byte                                 `json:"token"`
	BurnValue uint64                                 `json:"burnValue"`
	ExtraType byte                                   `json:"extraType"`
	ExtraData []byte                                 `json:"extraData"`
	Statement *json_Only_TransactionPayloadStatement `json:"statement"`
	Proof     *json_Only_TransactionPayloadProof     `json:"proof"`
}

type json_TransactionZether struct {
	*json_Transaction
	*json_Only_TransactionZether
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

			statementJson := &json_Only_TransactionPayloadStatement{
				payload.Statement.RingSize,
				helpers.ConvertBN256Array(payload.Statement.CLn),
				helpers.ConvertBN256Array(payload.Statement.CRn),
				helpers.ConvertBN256Array(payload.Statement.Publickeylist),
				helpers.ConvertBN256Array(payload.Statement.C),
				payload.Statement.D.EncodeCompressed(),
				payload.Statement.Fees,
				payload.Statement.Roothash,
			}

			proofJson := &json_Only_TransactionPayloadProof{
				payload.Proof.BA.EncodeCompressed(),
				payload.Proof.BS.EncodeCompressed(),
				payload.Proof.A.EncodeCompressed(),
				payload.Proof.B.EncodeCompressed(),
				helpers.ConvertBN256Array(payload.Proof.CLnG),
				helpers.ConvertBN256Array(payload.Proof.CRnG),
				helpers.ConvertBN256Array(payload.Proof.C_0G),
				helpers.ConvertBN256Array(payload.Proof.DG),
				helpers.ConvertBN256Array(payload.Proof.Y_0G),
				helpers.ConvertBN256Array(payload.Proof.GG),
				helpers.ConvertBN256Array(payload.Proof.C_XG),
				helpers.ConvertBN256Array(payload.Proof.Y_XG),
				payload.Proof.U,
				payload.Proof.U1,
				payload.Proof.F,
				payload.Proof.Z_A,
				payload.Proof.T_1.EncodeCompressed(),
				payload.Proof.T_2.EncodeCompressed(),
				payload.Proof.That,
				payload.Proof.Mu,
				payload.Proof.C,
				payload.Proof.S_sk,
				payload.Proof.S_r,
				payload.Proof.S_b,
				payload.Proof.S_tau,
				payload.Proof.Ip,
			}

			payloadsJson[i] = &json_Only_TransactionPayload{
				payload.Token,
				payload.BurnValue,
				payload.ExtraType,
				payload.ExtraData,
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
			return errors.New("Invalid base.TxScript")
		}

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
