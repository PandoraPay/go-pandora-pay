package transaction

import (
	"encoding/json"
	"errors"
	"math"
	"pandora-pay/blockchain/data_storage/assets/asset"
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
	PublicKeyIndex        byte             `json:"publicKeyIndex"`
	RegistrationSignature helpers.HexBytes `json:"signature"`
}

type json_TransactionDataDelegatedStakingUpdate struct {
	DelegatedStakingHasNewInfo   bool             `json:"delegatedStakingHasNewInfo"`
	DelegatedStakingNewPublicKey helpers.HexBytes `json:"delegatedStakingNewPublicKey"` //20 byte
	DelegatedStakingNewFee       uint64           `json:"delegatedStakingNewFee"`
}

type json_Transaction struct {
	Version    transaction_type.TransactionVersion `json:"version"`
	Size       uint64                              `json:"size"`
	SpaceExtra uint64                              `json:"spaceExtra"`
	Hash       helpers.HexBytes                    `json:"hash"`
}

type json_TransactionSimple struct {
	*json_Transaction
	TxScript    transaction_simple.ScriptType           `json:"txScript"`
	DataVersion transaction_data.TransactionDataVersion `json:"dataVersion"`
	Data        helpers.HexBytes                        `json:"data"`
	Nonce       uint64                                  `json:"nonce"`
	Fee         uint64                                  `json:"fee"`
	FeeVersion  bool                                    `json:"feeVersion"`
	Vin         *json_TransactionSimpleInput            `json:"vin"`
	Extra       interface{}                             `json:"extra"`
}

type json_TransactionSimpleInput struct {
	PublicKey helpers.HexBytes `json:"publicKey,omitempty"` //32
	Signature helpers.HexBytes `json:"signature"`           //64
}

type json_Only_TransactionSimpleExtraUpdateDelegate struct {
	DelegatedStakingClaimAmount uint64                                      `json:"delegatedStakingClaimAmount"`
	DelegatedStakingUpdate      *json_TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
}

type json_Only_TransactionSimpleExtraUnstake struct {
	Amount uint64 `json:"amount"`
}

type json_Only_TransactionZether struct {
	ChainHeight uint64                          `json:"chainHeight"`
	ChainHash   helpers.HexBytes                `json:"chainHash"`
	Payloads    []*json_Only_TransactionPayload `json:"payloads"`
}

type json_Only_TransactionZetherPayloadExtraDelegateStake struct {
	DelegatePublicKey      helpers.HexBytes                            `json:"delegatePublicKey"`
	ConvertToUnclaimed     bool                                        `json:"convertToUnclaimed"`
	DelegatedStakingUpdate *json_TransactionDataDelegatedStakingUpdate `json:"delegatedStakingUpdate"`
	DelegateSignature      helpers.HexBytes                            `json:"delegateSignature"`
}

type json_Only_TransactionZetherPayloadExtraClaim struct {
	DelegatePublicKey           helpers.HexBytes `json:"delegatePublicKey"`
	DelegatedStakingClaimAmount uint64           `json:"delegatedStakingClaimAmount"`
	RegistrationIndex           byte             `json:"registrationIndex"`
	DelegateSignature           helpers.HexBytes `json:"delegateSignature"`
}

type json_Only_TransactionZetherPayloadExtraAssetCreate struct {
	Asset *asset.Asset `json:"asset"`
}

type json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease struct {
	AssetId              helpers.HexBytes
	ReceiverPublicKey    helpers.HexBytes //must be registered before
	Value                uint64
	AssetSupplyPublicKey helpers.HexBytes //TODO: it can be bloomed
	AssetSignature       helpers.HexBytes
}

type json_Only_TransactionZetherStatement struct {
	RingSize      uint64             `json:"ringSize"`
	CLn           []helpers.HexBytes `json:"cLn"`
	CRn           []helpers.HexBytes `json:"cRn"`
	Publickeylist []helpers.HexBytes `json:"publickeylist"`
	C             []helpers.HexBytes `json:"c"`
	D             helpers.HexBytes   `json:"d"`
	Fee           uint64             `json:"fee"`
}

type json_Only_TransactionPayload struct {
	PayloadScript   transaction_zether_payload.PayloadScriptType `json:"payloadScript"`
	Asset           helpers.HexBytes                             `json:"asset"`
	BurnValue       uint64                                       `json:"burnValue"`
	DataVersion     transaction_data.TransactionDataVersion      `json:"dataType"`
	Data            helpers.HexBytes                             `json:"data"`
	Registrations   []*json_TransactionDataRegistration          `json:"registrations"`
	Statement       *json_Only_TransactionZetherStatement        `json:"statement"`
	FeeRate         uint64                                       `json:"feeRate"`
	FeeLeadingZeros byte                                         `json:"feeLeadingZeros"`
	Proof           helpers.HexBytes                             `json:"proof"`
	Extra           interface{}                                  `json:"extra"`
}

type json_TransactionZether struct {
	*json_Transaction
	*json_Only_TransactionZether
}

func (tx *Transaction) MarshalJSON() ([]byte, error) {

	txJson := &json_Transaction{
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
			base.FeeVersion,
			vinJson,
			nil,
		}

		switch base.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUpdateDelegate)
			simpleJson.Extra = &json_Only_TransactionSimpleExtraUpdateDelegate{
				extra.DelegatedStakingClaimAmount,
				&json_TransactionDataDelegatedStakingUpdate{
					extra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
					extra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
					extra.DelegatedStakingUpdate.DelegatedStakingNewFee,
				},
			}
		case transaction_simple.SCRIPT_UNSTAKE:
			extra := base.Extra.(*transaction_simple_extra.TransactionSimpleExtraUnstake)
			simpleJson.Extra = json_Only_TransactionSimpleExtraUnstake{
				extra.Amount,
			}
		default:
			return nil, errors.New("Invalid simple.TxScript")
		}

		return json.Marshal(simpleJson)

	case transaction_type.TX_ZETHER:
		base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

		payloadsJson := make([]*json_Only_TransactionPayload, len(base.Payloads))
		for i, payload := range base.Payloads {

			registrations := make([]*json_TransactionDataRegistration, len(payload.Registrations.Registrations))
			for i, reg := range payload.Registrations.Registrations {
				registrations[i] = &json_TransactionDataRegistration{
					reg.PublicKeyIndex,
					reg.RegistrationSignature,
				}
			}

			statementJson := &json_Only_TransactionZetherStatement{
				payload.Statement.RingSize,
				helpers.ConvertBN256Array(payload.Statement.CLn),
				helpers.ConvertBN256Array(payload.Statement.CRn),
				helpers.ConvertBN256Array(payload.Statement.Publickeylist),
				helpers.ConvertBN256Array(payload.Statement.C),
				payload.Statement.D.EncodeCompressed(),
				payload.Statement.Fee,
			}

			w := helpers.NewBufferWriter()
			payload.Proof.Serialize(w)
			proofJson := w.Bytes()

			var extra interface{}

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
				//no payload
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake)
				extra = &json_Only_TransactionZetherPayloadExtraDelegateStake{
					payloadExtra.DelegatePublicKey,
					payloadExtra.ConvertToUnclaimed,
					&json_TransactionDataDelegatedStakingUpdate{
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
						payloadExtra.DelegatedStakingUpdate.DelegatedStakingNewFee,
					},
					payloadExtra.DelegateSignature,
				}
			case transaction_zether_payload.SCRIPT_CLAIM:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim)
				extra = &json_Only_TransactionZetherPayloadExtraClaim{
					payloadExtra.DelegatePublicKey,
					payloadExtra.DelegatedStakingClaimAmount,
					payloadExtra.RegistrationIndex,
					payloadExtra.DelegateSignature,
				}
			case transaction_zether_payload.SCRIPT_ASSET_CREATE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate)
				extra = &json_Only_TransactionZetherPayloadExtraAssetCreate{
					payloadExtra.Asset,
				}
			case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				payloadExtra := payload.Extra.(*transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease)
				extra = &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{
					payloadExtra.AssetId,
					payloadExtra.ReceiverPublicKey,
					payloadExtra.Value,
					payloadExtra.AssetSignature,
					payloadExtra.AssetSupplyPublicKey,
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
				registrations,
				statementJson,
				payload.FeeRate,
				payload.FeeLeadingZeros,
				proofJson,
				extra,
			}

		}

		zetherJson := &json_TransactionZether{
			txJson,
			&json_Only_TransactionZether{
				base.ChainHeight,
				base.ChainHash,
				payloadsJson,
			},
		}

		return json.Marshal(zetherJson)

	default:
		return nil, errors.New("Invalid Tx Version")
	}

}

func (tx *Transaction) UnmarshalJSON(data []byte) (err error) {

	txOnlyJson := &json_Transaction{}
	if err = json.Unmarshal(data, txOnlyJson); err != nil {
		return
	}

	switch txOnlyJson.Version {
	case transaction_type.TX_SIMPLE, transaction_type.TX_ZETHER:
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
			simpleJson.FeeVersion,
			vin,
			nil,
		}
		tx.TransactionBaseInterface = base

		switch simpleJson.TxScript {
		case transaction_simple.SCRIPT_UPDATE_DELEGATE:

			extraJson := &json_Only_TransactionSimpleExtraUpdateDelegate{}
			if err = json.Unmarshal(data, extraJson); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUpdateDelegate{
				nil,
				extraJson.DelegatedStakingClaimAmount,
				&transaction_data.TransactionDataDelegatedStakingUpdate{
					extraJson.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
					extraJson.DelegatedStakingUpdate.DelegatedStakingNewFee,
				},
			}

		case transaction_simple.SCRIPT_UNSTAKE:
			extraJSON := &json_Only_TransactionSimpleExtraUnstake{}
			if err = json.Unmarshal(data, extraJSON); err != nil {
				return
			}

			base.Extra = &transaction_simple_extra.TransactionSimpleExtraUnstake{
				Amount: extraJSON.Amount,
			}
		default:
			return errors.New("Invalid json Simple TxScript")
		}

	case transaction_type.TX_ZETHER:

		simpleZether := &json_Only_TransactionZether{}
		if err = json.Unmarshal(data, simpleZether); err != nil {
			return
		}

		payloads := make([]*transaction_zether_payload.TransactionZetherPayload, len(simpleZether.Payloads))
		for i, payload := range simpleZether.Payloads {

			statement := &crypto.Statement{
				RingSize: payload.Statement.RingSize,
				Fee:      payload.Statement.Fee,
			}

			if statement.CLn, err = helpers.ConvertToBN256Array(payload.Statement.CLn); err != nil {
				return
			}
			if statement.CRn, err = helpers.ConvertToBN256Array(payload.Statement.CRn); err != nil {
				return
			}
			if statement.Publickeylist, err = helpers.ConvertToBN256Array(payload.Statement.Publickeylist); err != nil {
				return
			}
			if statement.C, err = helpers.ConvertToBN256Array(payload.Statement.C); err != nil {
				return
			}

			statement.D = new(bn256.G1)
			if err = statement.D.DecodeCompressed(payload.Statement.D); err != nil {
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
				payload.PayloadScript,
				payload.Asset,
				payload.BurnValue,
				payload.DataVersion,
				payload.Data,
				&transaction_zether_registrations.TransactionZetherDataRegistrations{
					Registrations: make([]*transaction_zether_registrations.TransactionZetherDataRegistration, len(payload.Registrations)),
				},
				statement,
				payload.FeeRate,
				payload.FeeLeadingZeros,
				proof,
				nil,
			}

			for i, reg := range payload.Registrations {
				payloads[i].Registrations.Registrations[i] = &transaction_zether_registrations.TransactionZetherDataRegistration{
					reg.PublicKeyIndex,
					reg.RegistrationSignature,
				}
			}

			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_TRANSFER:
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE:
				extraJSON := &json_Only_TransactionZetherPayloadExtraDelegateStake{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraDelegateStake{
					nil,
					extraJSON.DelegatePublicKey,
					extraJSON.ConvertToUnclaimed,
					&transaction_data.TransactionDataDelegatedStakingUpdate{
						extraJSON.DelegatedStakingUpdate.DelegatedStakingHasNewInfo,
						extraJSON.DelegatedStakingUpdate.DelegatedStakingNewPublicKey,
						extraJSON.DelegatedStakingUpdate.DelegatedStakingNewFee,
					},
					extraJSON.DelegateSignature,
				}

			case transaction_zether_payload.SCRIPT_CLAIM:
				extraJSON := &json_Only_TransactionZetherPayloadExtraClaim{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}

				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraClaim{
					nil,
					extraJSON.DelegatePublicKey,
					extraJSON.DelegatedStakingClaimAmount,
					extraJSON.RegistrationIndex,
					extraJSON.DelegateSignature,
				}

			case transaction_zether_payload.SCRIPT_ASSET_CREATE:
				extraJSON := &json_Only_TransactionZetherPayloadExtraAssetCreate{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetCreate{
					Asset: extraJSON.Asset,
				}
			case transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				extraJSON := &json_Only_TransactionZetherPayloadExtraAssetSupplyIncrease{}
				if err := json.Unmarshal(data, extraJSON); err != nil {
					return err
				}
				payloads[i].Extra = &transaction_zether_payload_extra.TransactionZetherPayloadExtraAssetSupplyIncrease{
					nil,
					extraJSON.AssetId,
					extraJSON.ReceiverPublicKey,
					extraJSON.Value,
					extraJSON.AssetSignature,
					extraJSON.AssetSupplyPublicKey,
				}

			default:
				return errors.New("Invalid Zether TxScript")
			}

		}

		base := &transaction_zether.TransactionZether{
			ChainHeight: simpleZether.ChainHeight,
			ChainHash:   simpleZether.ChainHash,
			Payloads:    payloads,
		}

		tx.TransactionBaseInterface = base

	default:
		return errors.New("Invalid Version")
	}

	return nil
}
