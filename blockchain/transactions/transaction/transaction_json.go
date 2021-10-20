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
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload/transaction_zether_payload_extra"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type json_TransactionDataRegistration struct {
	PublicKeyIndex        uint64           `json:"publicKeyIndex"`
	RegistrationSignature helpers.HexBytes `json:"signature"`
}

type json_TransactionDataDelegatedStakingUpdate struct {
	DelegatedStakingHasNewInfo   bool             `json:"delegatedStakingHasNewInfo"`
	DelegatedStakingNewPublicKey helpers.HexBytes `json:"delegatedStakingNewPublicKey"` //20 byte
	DelegatedStakingNewFee       uint64           `json:"delegatedStakingNewFee"`
}

type json_Transaction struct {
	Version transaction_type.TransactionVersion `json:"version"`
	Size    uint64                              `json:"size"`
	Hash    helpers.HexBytes                    `json:"hash"`
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
	DelegatedStakingClaimAmount uint64                                      `json:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *json_TransactionDataDelegatedStakingUpdate `json:"DelegatedStakingUpdate"`
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

type json_Only_TransactionZether struct {
	Height        uint64                              `json:"height"`
	Registrations []*json_TransactionDataRegistration `json:"registrations"`
	Payloads      []*json_Only_TransactionPayload     `json:"payloads"`
}

type json_Only_TransactionZetherPayloadExtraDelegateStake struct {
	DelegatePublicKey            helpers.HexBytes `json:"delegatePublicKey"`
	DelegatedStakingNewInfo      bool             `json:"delegatedStakingNewInfo"`
	DelegatedStakingNewPublicKey helpers.HexBytes `json:"delegatedStakingNewPublicKey"`
	DelegatedStakingNewFee       uint64           `json:"delegatedStakingNewFee"`
	DelegateSignature            helpers.HexBytes `json:"delegateSignature"`
}

type json_Only_TransactionZetherPayloadExtraClaimStake struct {
	DelegatePublicKey           helpers.HexBytes `json:"delegatePublicKey"`
	DelegatedStakingClaimAmount uint64           `json:"delegatedStakingClaimAmount"`
	RegistrationIndex           byte             `json:"registrationIndex"`
	DelegateSignature           helpers.HexBytes `json:"delegateSignature"`
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
	PayloadScript transaction_zether_payload.PayloadScriptType `json:"payloadScript"`
	Asset         helpers.HexBytes                             `json:"asset"`
	BurnValue     uint64                                       `json:"burnValue"`
	DataVersion   transaction_data.TransactionDataVersion      `json:"dataType"`
	Data          helpers.HexBytes                             `json:"data"`
	Statement     *json_Only_TransactionZetherStatement        `json:"statement"`
	Proof         helpers.HexBytes                             `json:"proof"`
	Extra         interface{}                                  `json:"extra"`
}

type json_TransactionZether struct {
	*json_Transaction
	*json_Only_TransactionZether
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	txJson := &json_Transaction{
		tx.Version,
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
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleUpdateDelegate)
			return json.Marshal(&json_TransactionSimpleUpdateDelegate{
				simpleJson,
				&json_Only_TransactionSimpleUpdateDelegate{
					extra.DelegatedStakingClaimAmount,
					&json_TransactionDataDelegatedStakingUpdate{
						extra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
						extra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
						extra.DelegatedStakingUpdate.DelegatedStakingNewFee,
					},
				},
			})
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleUnstake)
			return json.Marshal(&json_TransactionSimpleUnstake{
				simpleJson,
				&json_Only_TransactionSimpleUnstake{
					extra.Amount,
				},
			})
		default:
			return nil, errors.New("Invalid simple.TxScript")
		}
	case transaction_type.TX_ZETHER:
		base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

		registrations := make([]*json_TransactionDataRegistration, len(base.Registrations.Registrations))
		for i, reg := range base.Registrations.Registrations {
			registrations[i] = &json_TransactionDataRegistration{
				reg.PublicKeyIndex,
				reg.RegistrationSignature,
			}
		}

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

			var extra interface{}

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
				//no payload
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadDelegateStake)
				extra = &json_Only_TransactionZetherPayloadExtraDelegateStake{
					payloadExtra.DelegatePublicKey,
					payloadExtra.DelegatedStakingNewInfo,
					payloadExtra.DelegatedStakingNewPublicKey,
					payloadExtra.DelegatedStakingNewFee,
					payloadExtra.DelegateSignature,
				}
			case transaction_zether_payload.SCRIPT_CLAIM_STAKE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadClaimStake)
				extra = &json_Only_TransactionZetherPayloadExtraClaimStake{
					payloadExtra.DelegatePublicKey,
					payloadExtra.DelegatedStakingClaimAmount,
					payloadExtra.RegistrationIndex,
					payloadExtra.DelegateSignature,
				}
			default:
				return nil, errors.New("Invalid zether.TxScript")
			}

			payloadsJson[i] = &json_Only_TransactionPayload{
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				payload.DataVersion,
				payload.Data,
				statementJson,
				proofJson,
				extra,
			}

		}

		zetherJson := &json_TransactionZether{
			txJson,
			&json_Only_TransactionZether{
				base.Height,
				registrations,
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

			base.Extra = &transaction_simple_extra.TransactionSimpleUpdateDelegate{
				DelegatedStakingClaimAmount: extraJson.DelegatedStakingClaimAmount,
				DelegatedStakingUpdate: &transaction_data.TransactionDataDelegatedStakingUpdate{
					extraJson.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewFee,
				},
			}

		case transaction_simple.SCRIPT_UNSTAKE:
			extraJSON := &json_Only_TransactionSimpleUnstake{}
			if err := json.Unmarshal(data, extraJSON); err != nil {
				return err
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleUnstake{
				Amount: extraJSON.Amount,
			}
		default:
			return errors.New("Invalid json Simple TxScript")
		}

	case transaction_type.TX_ZETHER:

		simpleZether := &json_Only_TransactionZether{}
		if err := json.Unmarshal(data, simpleZether); err != nil {
			return err
		}

		payloads := make([]*transaction_zether_payload.TransactionZetherPayload, len(simpleZether.Payloads))
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

			payloads[i] = &transaction_zether_payload.TransactionZetherPayload{
				Asset:       payload.Asset,
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

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				extraJSON := &json_Only_TransactionZetherPayloadExtraDelegateStake{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadDelegateStake{
					DelegatePublicKey:            extraJSON.DelegatePublicKey,
					DelegatedStakingNewInfo:      extraJSON.DelegatedStakingNewInfo,
					DelegatedStakingNewPublicKey: extraJSON.DelegatedStakingNewPublicKey,
					DelegatedStakingNewFee:       extraJSON.DelegatedStakingNewFee,
					DelegateSignature:            extraJSON.DelegateSignature,
				}

			case transaction_zether_payload.SCRIPT_CLAIM_STAKE:
				extraJSON := &json_Only_TransactionZetherPayloadExtraClaimStake{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadClaimStake{
					DelegatePublicKey:           extraJSON.DelegatePublicKey,
					RegistrationIndex:           extraJSON.RegistrationIndex,
					DelegateSignature:           extraJSON.DelegateSignature,
					DelegatedStakingClaimAmount: extraJSON.DelegatedStakingClaimAmount,
				}

			default:
				return errors.New("Invalid Zether TxScript")
			}

		}

		base := &transaction_zether.TransactionZether{
			Height: simpleZether.Height,
			Registrations: &transaction_zether_registrations.TransactionZetherDataRegistrations{
				Registrations: make([]*transaction_zether_registrations.TransactionZetherDataRegistration, len(simpleZether.Registrations)),
			},
			Payloads: payloads,
		}

		for i, reg := range simpleZether.Registrations {
			base.Registrations.Registrations[i] = &transaction_zether_registrations.TransactionZetherDataRegistration{
				reg.PublicKeyIndex,
				reg.RegistrationSignature,
			}
		}

		tx.TransactionBaseInterface = base

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
